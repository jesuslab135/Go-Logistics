// Package scheduler runs the reports on a clock: it decides what is due,
// claims the run, builds and sends the report, and records what happened.
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/platform/mail"
	"fleet/internal/platform/pdf"
	"fleet/internal/platform/reports"
)

// Run statuses, as stored in report_run.status.
const (
	StatusSent         = "sent"
	StatusNotSent      = "not_sent"
	StatusEmpty        = "skipped_empty"
	StatusNoRecipients = "skipped_no_recipients"
	StatusFailed       = "failed"
	// StatusNotClaimed is returned, never stored: the run belongs to someone
	// else or is already finished.
	StatusNotClaimed = "not_claimed"
)

const (
	// maxAttempts is how many times a failing report is tried before it is
	// left as failed for a person to look at.
	maxAttempts = 3
	// staleMinutes is how long a run may stay "running" before its owner is
	// presumed dead. No report takes this long.
	staleMinutes = 30
	// lockKey is the advisory lock one tick holds. The number is arbitrary;
	// it only has to be the same in every process.
	lockKey int64 = 7205001
)

// Store is the slice of the generated queries the runner needs.
// *gen.Queries satisfies it.
type Store interface {
	reports.Querier
	ListReportCompanies(ctx context.Context) ([]gen.ListReportCompaniesRow, error)
	ListActiveReportRecipientEmails(ctx context.Context, arg gen.ListActiveReportRecipientEmailsParams) ([]string, error)
	ClaimReportRun(ctx context.Context, arg gen.ClaimReportRunParams) (gen.ClaimReportRunRow, error)
	FinishReportRun(ctx context.Context, arg gen.FinishReportRunParams) (int64, error)
}

// LogoLoader returns a company's logo bytes, or nil when it has none or it
// cannot be read. A missing logo never fails a report.
type LogoLoader func(ctx context.Context, c reports.Company) []byte

type Deps struct {
	// Pool is used only to take the advisory lock. Nil skips locking, which is
	// right for the CLI and for unit tests.
	Pool  *pgxpool.Pool
	Store Store
	// Mail is nil when no mail server is configured.
	Mail     mail.Sender
	Logo     LogoLoader
	Logger   *slog.Logger
	Now      func() time.Time
	SendHour int
}

type Runner struct {
	d Deps
}

func NewRunner(d Deps) *Runner {
	if d.Logger == nil {
		d.Logger = slog.Default()
	}
	if d.Now == nil {
		d.Now = time.Now
	}
	if d.Logo == nil {
		d.Logo = func(context.Context, reports.Company) []byte { return nil }
	}
	return &Runner{d: d}
}

// Tick runs every report that is due, for every company. It returns an error
// only when it could not start; a report that fails is recorded and the loop
// goes on to the next.
func (r *Runner) Tick(ctx context.Context) error {
	release, ok, err := r.lock(ctx)
	if err != nil {
		return fmt.Errorf("scheduler: advisory lock: %w", err)
	}
	if !ok {
		r.d.Logger.Info("scheduler: another process holds the tick, skipping")
		return nil
	}
	defer release()

	companies, err := r.d.Store.ListReportCompanies(ctx)
	if err != nil {
		return fmt.Errorf("scheduler: list companies: %w", err)
	}

	now := r.d.Now()
	for _, row := range companies {
		c := reports.Company{ID: row.ID, Name: row.Name, Logo: row.Logo, Timezone: row.Timezone, Currency: row.Currency}
		loc, err := reports.LoadLocation(c.Timezone)
		if err != nil {
			r.d.Logger.Warn("scheduler: company timezone", "company_id", c.ID, "error", err)
		}
		for _, kind := range reports.Kinds {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			period := reports.LatestDue(kind, now, loc, r.d.SendHour)
			status, err := r.RunOne(ctx, c, kind, period, false)
			if status == StatusNotClaimed {
				continue
			}
			r.d.Logger.Info("scheduler: report run",
				"company_id", c.ID, "report", kind, "period_start", period.StartDate().Format("2006-01-02"),
				"status", status, "error", err)
		}
	}
	return nil
}

// lock takes the advisory lock on a connection of its own. Advisory locks
// belong to the session, so the connection is held until release.
func (r *Runner) lock(ctx context.Context) (release func(), ok bool, err error) {
	if r.d.Pool == nil {
		return func() {}, true, nil
	}
	conn, err := r.d.Pool.Acquire(ctx)
	if err != nil {
		return nil, false, err
	}
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", lockKey).Scan(&ok); err != nil || !ok {
		conn.Release()
		return nil, false, err
	}
	return func() {
		// A fresh context: the caller's may already be cancelled, and the
		// lock must be given back regardless.
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := conn.Exec(unlockCtx, "SELECT pg_advisory_unlock($1)", lockKey); err != nil {
			// Closing the session releases the lock; returning the
			// connection to the pool would keep it held.
			_ = conn.Conn().Close(unlockCtx)
		}
		conn.Release()
	}, true, nil
}

// RunOne claims one company's report for one period and delivers it. The
// returned status is what was stored, or StatusNotClaimed when there was
// nothing to do. force re-sends a report that was already handled.
func (r *Runner) RunOne(ctx context.Context, c reports.Company, kind reports.Kind, p reports.Period, force bool) (string, error) {
	claim, err := r.d.Store.ClaimReportRun(ctx, gen.ClaimReportRunParams{
		CompanyID:    c.ID,
		ReportKind:   string(kind),
		PeriodStart:  p.StartDate(),
		Force:        force,
		MaxAttempts:  maxAttempts,
		StaleMinutes: staleMinutes,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return StatusNotClaimed, nil
	}
	if err != nil {
		return StatusFailed, fmt.Errorf("scheduler: claim run: %w", err)
	}

	status, recipients, runErr := r.deliver(ctx, c, kind, p)

	finish := gen.FinishReportRunParams{
		ID:         claim.ID,
		Attempts:   claim.Attempts,
		Status:     status,
		Recipients: int32(recipients),
	}
	if runErr != nil {
		msg := runErr.Error()
		finish.Error = &msg
	}
	// A fresh context: if the run failed because ctx was cancelled, the
	// failure must still be written, or the row stays "running" for 30 minutes.
	finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	rows, err := r.d.Store.FinishReportRun(finishCtx, finish)
	if err != nil {
		return status, errors.Join(runErr, fmt.Errorf("scheduler: record run: %w", err))
	}
	if rows == 0 {
		// Someone else's forced or stale-timeout re-claim took this run over
		// while we were working. Their result is what is stored; ours must
		// not be reported as an error, since delivery itself may well have
		// succeeded.
		r.d.Logger.Warn("scheduler: run was taken over before it finished; result not recorded",
			"company_id", c.ID, "report", kind, "period_start", p.StartDate().Format("2006-01-02"), "status", status)
	}
	return status, runErr
}

func (r *Runner) deliver(ctx context.Context, c reports.Company, kind reports.Kind, p reports.Period) (status string, recipients int, err error) {
	// A panic in one report (a rendering bug on one company's data) must not
	// take the API process down with it.
	defer func() {
		if rec := recover(); rec != nil {
			status, err = StatusFailed, fmt.Errorf("scheduler: panic: %v", rec)
		}
	}()

	to, err := r.d.Store.ListActiveReportRecipientEmails(ctx, gen.ListActiveReportRecipientEmailsParams{
		CompanyID:  c.ID,
		ReportKind: string(kind),
	})
	if err != nil {
		return StatusFailed, 0, fmt.Errorf("list recipients: %w", err)
	}

	doc, err := pdf.Render(ctx, r.d.Store, kind, c, p, r.d.Logo(ctx, c), r.d.Now())
	if err != nil {
		return StatusFailed, len(to), fmt.Errorf("render: %w", err)
	}
	if doc.Empty {
		return StatusEmpty, len(to), nil
	}
	if len(to) == 0 {
		return StatusNoRecipients, 0, nil
	}
	if r.d.Mail == nil {
		return StatusNotSent, len(to), nil
	}

	loc := p.Start.Location()
	msg := mail.Message{
		To:      to,
		Subject: fmt.Sprintf("%s — %s", kind.Title(), c.Name),
		Body: fmt.Sprintf("Adjunto encontrará el %s de %s.\n\nPeriodo: %s a %s.\n\nEste mensaje se genera automáticamente.\n",
			lowerFirst(kind.Title()), c.Name, pdf.Date(p.Start, loc), pdf.Date(p.LastDay(), loc)),
		Attachments: []mail.Attachment{{Filename: doc.Filename, ContentType: "application/pdf", Data: doc.Data}},
	}
	if err := r.d.Mail.Send(ctx, msg); err != nil {
		return StatusFailed, len(to), err
	}
	return StatusSent, len(to), nil
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'A' && r[0] <= 'Z' {
		r[0] += 'a' - 'A'
	}
	return string(r)
}
