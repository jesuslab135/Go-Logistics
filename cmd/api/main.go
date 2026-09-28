package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // the distroless image carries no timezone database

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"fleet/internal/auth"
	"fleet/internal/config"
	"fleet/internal/db"
	"fleet/internal/db/gen"
	"fleet/internal/http/handler"
	"fleet/internal/platform/dbctx"
	"fleet/internal/platform/mail"
	"fleet/internal/platform/scheduler"
	"fleet/internal/platform/storage"

	_ "fleet/docs" // generated OpenAPI spec (swag init)
)

// @title						Fleet API
// @version					1.0
// @description				Go port of the fleet-management backend (auth, companies, uploads).
// @BasePath					/
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	_ = godotenv.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("server terminated", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	queries := gen.New(dbctx.New(pool))
	tokens := auth.NewTokenService(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	blobs, err := storage.FromEnv()
	if err != nil {
		return err
	}

	router := handler.NewRouter(handler.Deps{
		Queries:     queries,
		Pool:        pool,
		Tokens:      tokens,
		Verifier:    handler.NewEmployeeCredentialVerifier(queries),
		Storage:     blobs,
		Logger:      logger,
		CORSOrigins: cfg.CORSOrigins,
		Production:  cfg.IsProduction(),
		Swagger:     cfg.Swagger,
	})

	stopScheduler, err := startScheduler(cfg, pool, queries, blobs, logger)
	if err != nil {
		return err
	}
	// Deferred after pool.Close(), so it runs before it: a report in flight is
	// still reading from the pool, and stopScheduler waits for it.
	defer stopScheduler()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Shutdown closes the listeners first, so ListenAndServe returns as soon as
	// shutdown *starts*. Returning there would fire the deferred pool.Close()
	// underneath requests that are still running — the 15s budget below would
	// never elapse, and every deploy would abort in-flight transactions. So run
	// waits for Shutdown to report that the connections are idle.
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		logger.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
		}
	}()

	logger.Info("http server listening", "addr", cfg.HTTPAddr, "env", cfg.Env)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	<-shutdownDone
	return nil
}

// startScheduler starts the report scheduler when REPORTS_ENABLED says so, and
// returns the function that stops it. When it is off the returned function
// does nothing, so the caller does not have to know.
func startScheduler(cfg config.Config, pool *pgxpool.Pool, queries *gen.Queries, blobs storage.Storage, logger *slog.Logger) (func(), error) {
	if !cfg.Reports.Enabled {
		logger.Info("scheduler: off (REPORTS_ENABLED is not true)")
		return func() {}, nil
	}

	var sender mail.Sender
	if cfg.SMTP.Host != "" {
		sender = mail.NewSMTP(mail.Config{
			Host:     cfg.SMTP.Host,
			Port:     cfg.SMTP.Port,
			User:     cfg.SMTP.User,
			Password: cfg.SMTP.Password,
			From:     cfg.SMTP.From,
		})
	} else {
		logger.Warn("scheduler: SMTP_HOST is empty, reports will be built and recorded as not_sent")
	}

	runner := scheduler.NewRunner(scheduler.Deps{
		Pool:     pool,
		Store:    queries,
		Mail:     sender,
		Logo:     handler.NewLogoLoader(blobs, logger),
		Logger:   logger,
		SendHour: cfg.Reports.SendHour,
	})
	return scheduler.Start(runner, cfg.Reports.Tick, logger)
}
