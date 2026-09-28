package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"
	_ "time/tzdata" // the distroless image may carry no timezone database

	"fleet/internal/config"
	"fleet/internal/db"
	"fleet/internal/db/gen"
	"fleet/internal/http/handler"
	"fleet/internal/platform/mail"
	"fleet/internal/platform/reports"
	"fleet/internal/platform/scheduler"
	"fleet/internal/platform/storage"
)

// reportsCmd sends one company's report for one period, now.
//
// It is how a report recorded as not_sent is delivered once a mailbox is
// configured, and how a failed one is retried by hand. It goes through the
// same claim as the scheduler, so without --force it cannot send a report
// that was already sent.
func reportsCmd(args []string) error {
	if len(args) == 0 || args[0] != "send" {
		return fmt.Errorf("usage: fleet-cli reports send --company <id> --report <kind> [--period <date>] [--force]")
	}

	fs := flag.NewFlagSet("reports send", flag.ExitOnError)
	companyID := fs.Int64("company", 0, "company id")
	report := fs.String("report", "", "fuel_weekly or maintenance_monthly")
	periodArg := fs.String("period", "", "any date inside the period, YYYY-MM-DD (default: the last full period)")
	force := fs.Bool("force", false, "send even if this period was already handled")
	_ = fs.Parse(args[1:])

	if *companyID < 1 {
		return fmt.Errorf("--company is required")
	}
	kind, ok := reports.ParseKind(*report)
	if !ok {
		return fmt.Errorf("--report must be fuel_weekly or maintenance_monthly")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	q := gen.New(pool)

	row, err := q.GetReportCompany(ctx, *companyID)
	if err != nil {
		return fmt.Errorf("no company %d: %w", *companyID, err)
	}
	company := reports.Company{ID: row.ID, Name: row.Name, Logo: row.Logo, Timezone: row.Timezone, Currency: row.Currency}
	loc, err := reports.LoadLocation(company.Timezone)
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning:", err)
	}

	now := time.Now()
	period := reports.PeriodOf(kind, now, loc).Previous(kind)
	if *periodArg != "" {
		at, err := time.ParseInLocation("2006-01-02", *periodArg, loc)
		if err != nil {
			return fmt.Errorf("--period must be a date, YYYY-MM-DD")
		}
		period = reports.PeriodOf(kind, at, loc)
	}
	if !period.End.Before(now) && !period.End.Equal(now) {
		return fmt.Errorf("the period %s to %s has not ended yet",
			period.Start.Format("2006-01-02"), period.LastDay().Format("2006-01-02"))
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	blobs, err := storage.FromEnv()
	if err != nil {
		return err
	}
	var sender mail.Sender
	if smtp := cfg.SMTP; smtp.Host != "" {
		sender = mail.NewSMTP(mail.Config{Host: smtp.Host, Port: smtp.Port, User: smtp.User, Password: smtp.Password, From: smtp.From})
	}

	// No Pool: the advisory lock guards a whole tick, and this is one run.
	// The claim on report_run is what keeps it from colliding with the scheduler.
	runner := scheduler.NewRunner(scheduler.Deps{
		Store:    q,
		Mail:     sender,
		Logo:     handler.NewLogoLoader(blobs, logger),
		Logger:   logger,
		SendHour: cfg.Reports.SendHour,
	})

	status, err := runner.RunOne(ctx, company, kind, period, *force)
	fmt.Printf("company %d, %s, period starting %s: %s\n",
		company.ID, kind, period.StartDate().Format("2006-01-02"), status)
	if status == scheduler.StatusNotClaimed {
		fmt.Println("this period was already handled; pass --force to send it again")
	}
	return err
}
