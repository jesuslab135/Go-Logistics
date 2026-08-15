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

	"github.com/joho/godotenv"

	"fleet/internal/auth"
	"fleet/internal/config"
	"fleet/internal/db"
	"fleet/internal/db/gen"
	"fleet/internal/http/handler"
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

	queries := gen.New(pool)
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

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
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
	return nil
}
