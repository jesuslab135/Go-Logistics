//go:build integration

package handler

// The scheduler against a real Postgres: send-once under concurrency, retry,
// takeover of a dead run, and which companies are owed a report.

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/platform/mail"
	"fleet/internal/platform/reports"
	"fleet/internal/platform/scheduler"
)

type countingMail struct {
	sent atomic.Int64
	err  error
}

func (m *countingMail) Send(context.Context, mail.Message) error {
	if m.err != nil {
		return m.err
	}
	m.sent.Add(1)
	return nil
}

func (f reportFixture) runner(sender mail.Sender, now time.Time) *scheduler.Runner {
	return scheduler.NewRunner(scheduler.Deps{
		Pool: f.pool, Store: f.q, Mail: sender, Now: func() time.Time { return now }, SendHour: 6,
	})
}

// seedWeek gives the fixture's company fuel in the week before now and one
// recipient, so a tick at now owes it a fuel report.
func (f reportFixture) seedWeek(t *testing.T, now time.Time) reports.Period {
	t.Helper()
	asset := f.asset(t, f.token, "Unidad 7")
	employee := createDirectorCandidate(t, f.srv, f.token, "sched")
	var vendor struct {
		ID int64 `json:"id"`
	}
	postJSON(t, f.srv+"/api/v1/vendors", f.token, map[string]any{"name": "Gasolinera"}, http.StatusCreated, &vendor)

	week := reports.LatestDue(reports.FuelWeekly, now, f.loc, 6)
	f.fuel(t, asset, employee, vendor.ID, week.Start.Add(time.Hour), "100", "2500", "0", nil)
	postJSON(t, f.srv+"/api/v1/reports/recipients", f.token,
		map[string]any{"report_kind": "fuel_weekly", "email": "flota@cliente.mx"}, http.StatusCreated, nil)
	return week
}

func (f reportFixture) run(t *testing.T, kind reports.Kind, week reports.Period) gen.ReportRun {
	t.Helper()
	rows, err := f.q.ListReportRuns(context.Background(), gen.ListReportRunsParams{CompanyID: f.company, Lim: 50})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if r.ReportKind == string(kind) && r.PeriodStart.Format("2006-01-02") == week.StartDate().Format("2006-01-02") {
			return r
		}
	}
	t.Fatalf("no %s run for %s among %+v", kind, week.StartDate().Format("2006-01-02"), rows)
	return gen.ReportRun{}
}

// Eight schedulers ticking at once, as after a deploy that briefly runs two
// containers: one email, one row.
func TestConcurrentTicksSendOnce(t *testing.T) {
	f := newReportFixture(t)
	now := time.Date(2026, 9, 28, 7, 0, 0, 0, f.loc)
	week := f.seedWeek(t, now)

	sender := &countingMail{}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := f.runner(sender, now).Tick(context.Background()); err != nil {
				t.Errorf("tick: %v", err)
			}
		}()
	}
	wg.Wait()
	// Ticks that lost the advisory lock returned at once; run again so that
	// every one of them has had its turn at the claim.
	for range 3 {
		if err := f.runner(sender, now).Tick(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	if got := sender.sent.Load(); got != 1 {
		t.Fatalf("emails sent = %d, want 1", got)
	}
	run := f.run(t, reports.FuelWeekly, week)
	if run.Status != scheduler.StatusSent || run.Attempts != 1 || run.Recipients != 1 {
		t.Fatalf("run = %+v", run)
	}
}

func TestFailedRunIsRetriedThreeTimes(t *testing.T) {
	f := newReportFixture(t)
	now := time.Date(2026, 9, 28, 7, 0, 0, 0, f.loc)
	week := f.seedWeek(t, now)

	broken := &countingMail{err: errors.New("421 service not available")}
	for range 5 {
		if err := f.runner(broken, now).Tick(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	run := f.run(t, reports.FuelWeekly, week)
	if run.Status != scheduler.StatusFailed || run.Attempts != 3 {
		t.Fatalf("after five ticks: status = %s, attempts = %d, want failed and 3", run.Status, run.Attempts)
	}
	if run.Error == nil || *run.Error == "" {
		t.Fatal("a failed run must say why")
	}

	// The mail server comes back, but the attempts are spent: the scheduler
	// leaves it alone, and a person sends it with --force.
	working := &countingMail{}
	if err := f.runner(working, now).Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if working.sent.Load() != 0 {
		t.Fatal("a run with no attempts left was sent by the scheduler")
	}

	company := reports.Company{ID: f.company, Name: "Transportes Reporte", Timezone: f.loc.String(), Currency: "MXN"}
	status, err := f.runner(working, now).RunOne(context.Background(), company, reports.FuelWeekly, week, true)
	if err != nil || status != scheduler.StatusSent {
		t.Fatalf("forced run: status = %s, err = %v", status, err)
	}
}

// A run left "running" by a process that died is taken over once it is stale,
// and not before.
func TestStaleRunIsTakenOver(t *testing.T) {
	f := newReportFixture(t)
	now := time.Date(2026, 9, 28, 7, 0, 0, 0, f.loc)
	week := f.seedWeek(t, now)

	f.exec(t, `INSERT INTO report_run (company_id, report_kind, period_start, status, attempts, started_at)
	           VALUES ($1, 'fuel_weekly', $2, 'running', 1, now() - interval '5 minutes')`, f.company, week.StartDate())

	sender := &countingMail{}
	if err := f.runner(sender, now).Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if sender.sent.Load() != 0 {
		t.Fatal("a run that started five minutes ago was taken from its owner")
	}

	f.exec(t, `UPDATE report_run SET started_at = now() - interval '31 minutes' WHERE company_id = $1`, f.company)
	if err := f.runner(sender, now).Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if sender.sent.Load() != 1 {
		t.Fatalf("emails sent = %d, want 1 once the run is stale", sender.sent.Load())
	}
	if run := f.run(t, reports.FuelWeekly, week); run.Status != scheduler.StatusSent || run.Attempts != 2 {
		t.Fatalf("run = %+v", run)
	}
}

// A company of an inactive account is owed nothing.
func TestInactiveAccountIsSkipped(t *testing.T) {
	f := newReportFixture(t)
	now := time.Date(2026, 9, 28, 7, 0, 0, 0, f.loc)
	f.seedWeek(t, now)
	f.exec(t, `UPDATE account SET is_active = false WHERE id = (SELECT account_id FROM company WHERE id = $1)`, f.company)

	sender := &countingMail{}
	if err := f.runner(sender, now).Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if sender.sent.Load() != 0 {
		t.Fatal("a report was sent to a company of an inactive account")
	}
}

// A run whose process died on its last attempt can never be reclaimed
// (ClaimReportRun's own attempts < max_attempts predicate refuses it), so it
// would read "running" forever. Once it is stale, the tick closes it as
// failed instead, and sends nothing for it.
func TestStaleRunOnItsLastAttemptIsAbandoned(t *testing.T) {
	f := newReportFixture(t)
	now := time.Date(2026, 9, 28, 7, 0, 0, 0, f.loc)
	week := f.seedWeek(t, now)

	f.exec(t, `INSERT INTO report_run (company_id, report_kind, period_start, status, attempts, started_at)
	           VALUES ($1, 'fuel_weekly', $2, 'running', 3, now() - interval '31 minutes')`, f.company, week.StartDate())

	sender := &countingMail{}
	if err := f.runner(sender, now).Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if sender.sent.Load() != 0 {
		t.Fatal("a run on its last attempt was sent instead of abandoned")
	}
	run := f.run(t, reports.FuelWeekly, week)
	if run.Status != "failed" {
		t.Fatalf("status = %q, want failed", run.Status)
	}
	if run.Error == nil || *run.Error != "abandoned: the process stopped before the run finished" {
		t.Fatalf("error = %v, want the abandoned message", run.Error)
	}
}

// The same run, not yet stale, is left alone: its process may still be
// working on it.
func TestRunOnItsLastAttemptNotYetStaleStaysRunning(t *testing.T) {
	f := newReportFixture(t)
	now := time.Date(2026, 9, 28, 7, 0, 0, 0, f.loc)
	week := f.seedWeek(t, now)

	f.exec(t, `INSERT INTO report_run (company_id, report_kind, period_start, status, attempts, started_at)
	           VALUES ($1, 'fuel_weekly', $2, 'running', 3, now() - interval '5 minutes')`, f.company, week.StartDate())

	sender := &countingMail{}
	if err := f.runner(sender, now).Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if sender.sent.Load() != 0 {
		t.Fatal("a run on its last attempt was sent")
	}
	if run := f.run(t, reports.FuelWeekly, week); run.Status != "running" {
		t.Fatalf("status = %q, want it to stay running", run.Status)
	}
}
