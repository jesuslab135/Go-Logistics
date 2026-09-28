package scheduler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"fleet/internal/db/gen"
	"fleet/internal/platform/mail"
	"fleet/internal/platform/reports"
)

// fakeStore stands in for the database. Claims follow the rule ClaimReportRun
// implements in SQL: a period is handed out once, unless forced. attempts is
// tracked per run id (not per key): a forced claim reuses the same id and
// bumps its attempts, exactly as the SQL UPDATE does, so FinishReportRun can
// tell a stalled owner's stale attempts value apart from the current one.
type fakeStore struct {
	mu         sync.Mutex
	companies  []gen.ListReportCompaniesRow
	recipients map[int64][]string
	fuel       map[int64][]gen.ReportFuelByAssetRow
	fuelErr    map[int64]error
	claimed    map[string]int64 // key -> run id
	attempts   map[int64]int32  // run id -> current attempts
	finished   map[int64]gen.FinishReportRunParams
	nextID     int64
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		recipients: map[int64][]string{},
		fuel:       map[int64][]gen.ReportFuelByAssetRow{},
		fuelErr:    map[int64]error{},
		claimed:    map[string]int64{},
		attempts:   map[int64]int32{},
		finished:   map[int64]gen.FinishReportRunParams{},
	}
}

func (f *fakeStore) ListReportCompanies(context.Context) ([]gen.ListReportCompaniesRow, error) {
	return f.companies, nil
}

func (f *fakeStore) ListActiveReportRecipientEmails(_ context.Context, arg gen.ListActiveReportRecipientEmailsParams) ([]string, error) {
	return f.recipients[arg.CompanyID], nil
}

func (f *fakeStore) ClaimReportRun(_ context.Context, arg gen.ClaimReportRunParams) (gen.ClaimReportRunRow, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := fmt.Sprintf("%d|%s|%s", arg.CompanyID, arg.ReportKind, arg.PeriodStart.Format("2006-01-02"))
	if id, ok := f.claimed[key]; ok {
		if !arg.Force {
			return gen.ClaimReportRunRow{}, pgx.ErrNoRows
		}
		f.attempts[id]++
		return gen.ClaimReportRunRow{ID: id, Attempts: f.attempts[id]}, nil
	}
	f.nextID++
	id := f.nextID
	f.claimed[key] = id
	f.attempts[id] = 1
	return gen.ClaimReportRunRow{ID: id, Attempts: 1}, nil
}

// FinishReportRun mirrors the SQL's ownership check: it only records the
// result, and only reports one row affected, when arg.Attempts still matches
// the run's current attempts. A caller whose claim was superseded by a
// forced re-claim carries a stale value and affects no rows.
func (f *fakeStore) FinishReportRun(_ context.Context, arg gen.FinishReportRunParams) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if arg.Attempts != f.attempts[arg.ID] {
		return 0, nil
	}
	f.finished[arg.ID] = arg
	return 1, nil
}

func (f *fakeStore) ReportFuelByAsset(_ context.Context, arg gen.ReportFuelByAssetParams) ([]gen.ReportFuelByAssetRow, error) {
	if err := f.fuelErr[arg.CompanyID]; err != nil {
		return nil, err
	}
	return f.fuel[arg.CompanyID], nil
}

func (f *fakeStore) ReportMaintenanceByAsset(context.Context, gen.ReportMaintenanceByAssetParams) ([]gen.ReportMaintenanceByAssetRow, error) {
	return nil, nil
}

func (f *fakeStore) statuses() map[string]int {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := map[string]int{}
	for _, fin := range f.finished {
		out[fin.Status]++
	}
	return out
}

type fakeMail struct {
	mu   sync.Mutex
	sent []mail.Message
	err  error
}

func (f *fakeMail) Send(_ context.Context, m mail.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, m)
	return nil
}

func fuelRow(name string) gen.ReportFuelByAssetRow {
	return gen.ReportFuelByAssetRow{
		AssetID: 1, AssetName: name, MeterUnit: "km", VolumeUnit: "liters", FuelType: "diesel",
		Fills: 1, Volume: decimal.NewFromInt(100), Cost: decimal.NewFromInt(2500),
	}
}

var quiet = slog.New(slog.NewTextHandler(io.Discard, nil))

// monday is a Monday at 07:00 in Mexico City, an hour after the send hour.
func monday(t *testing.T) time.Time {
	t.Helper()
	mx, err := time.LoadLocation("America/Mexico_City")
	if err != nil {
		t.Fatal(err)
	}
	return time.Date(2026, 9, 28, 7, 0, 0, 0, mx)
}

func newTestRunner(t *testing.T, store *fakeStore, sender mail.Sender) *Runner {
	t.Helper()
	now := monday(t)
	return NewRunner(Deps{Store: store, Mail: sender, Logger: quiet, Now: func() time.Time { return now }, SendHour: 6})
}

func company(id int64, tz string) gen.ListReportCompaniesRow {
	return gen.ListReportCompaniesRow{ID: id, Name: "Transportes Durán", Timezone: tz, Currency: "MXN"}
}

func TestTickSendsOncePerPeriod(t *testing.T) {
	store := newFakeStore()
	store.companies = []gen.ListReportCompaniesRow{company(1, "America/Mexico_City")}
	store.recipients[1] = []string{"flota@cliente.mx", "dueno@cliente.mx"}
	store.fuel[1] = []gen.ReportFuelByAssetRow{fuelRow("Unidad 7")}
	sender := &fakeMail{}
	r := newTestRunner(t, store, sender)

	for range 3 {
		if err := r.Tick(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	if len(sender.sent) != 1 {
		t.Fatalf("emails sent = %d, want 1 after three ticks", len(sender.sent))
	}
	msg := sender.sent[0]
	if len(msg.To) != 2 {
		t.Fatalf("recipients = %v", msg.To)
	}
	if !strings.Contains(msg.Subject, "Reporte semanal de combustible") || !strings.Contains(msg.Subject, "Transportes Durán") {
		t.Fatalf("subject = %q", msg.Subject)
	}
	if !strings.Contains(msg.Body, "21/09/2026 a 27/09/2026") {
		t.Fatalf("body does not name the period: %q", msg.Body)
	}
	if len(msg.Attachments) != 1 || msg.Attachments[0].Filename != "combustible-semanal-2026-09-21.pdf" {
		t.Fatalf("attachments = %+v", msg.Attachments)
	}
	if !strings.HasPrefix(string(msg.Attachments[0].Data), "%PDF-") {
		t.Fatal("the attachment is not a PDF")
	}
}

func TestStatuses(t *testing.T) {
	tests := []struct {
		name       string
		recipients []string
		rows       []gen.ReportFuelByAssetRow
		sender     mail.Sender
		want       string
	}{
		{"sent", []string{"a@c.mx"}, []gen.ReportFuelByAssetRow{fuelRow("U")}, &fakeMail{}, StatusSent},
		{"no mail server", []string{"a@c.mx"}, []gen.ReportFuelByAssetRow{fuelRow("U")}, nil, StatusNotSent},
		{"no recipients", nil, []gen.ReportFuelByAssetRow{fuelRow("U")}, &fakeMail{}, StatusNoRecipients},
		{"no data", []string{"a@c.mx"}, nil, &fakeMail{}, StatusEmpty},
		{"mail server refuses", []string{"a@c.mx"}, []gen.ReportFuelByAssetRow{fuelRow("U")}, &fakeMail{err: errors.New("550 refused")}, StatusFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStore()
			store.recipients[1] = tt.recipients
			store.fuel[1] = tt.rows
			r := newTestRunner(t, store, tt.sender)

			c := reports.Company{ID: 1, Name: "C", Timezone: "America/Mexico_City", Currency: "MXN"}
			loc, _ := reports.LoadLocation(c.Timezone)
			p := reports.LatestDue(reports.FuelWeekly, monday(t), loc, 6)

			status, err := r.RunOne(context.Background(), c, reports.FuelWeekly, p, false)
			if status != tt.want {
				t.Fatalf("status = %s (err %v), want %s", status, err, tt.want)
			}
			if (tt.want == StatusFailed) != (err != nil) {
				t.Fatalf("err = %v for status %s", err, status)
			}
			got := store.finished[1]
			if got.Status != tt.want {
				t.Fatalf("stored status = %q, want %q", got.Status, tt.want)
			}
			if (got.Error != nil) != (tt.want == StatusFailed) {
				t.Fatalf("stored error = %v for status %s", got.Error, tt.want)
			}
		})
	}
}

// One company's failure is recorded and the others still get their report.
func TestOneCompanyFailingDoesNotStopTheRest(t *testing.T) {
	store := newFakeStore()
	for id := int64(1); id <= 3; id++ {
		store.companies = append(store.companies, company(id, "America/Mexico_City"))
		store.recipients[id] = []string{"a@c.mx"}
		store.fuel[id] = []gen.ReportFuelByAssetRow{fuelRow("U")}
	}
	store.fuelErr[2] = errors.New("connection reset")
	sender := &fakeMail{}

	if err := newTestRunner(t, store, sender).Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(sender.sent) != 2 {
		t.Fatalf("emails sent = %d, want 2", len(sender.sent))
	}
	if got := store.statuses(); got[StatusFailed] != 1 || got[StatusSent] != 2 {
		t.Fatalf("fuel statuses = %v", got)
	}
}

// A company whose timezone is unknown still gets its report, on the default.
func TestUnknownTimezoneFallsBack(t *testing.T) {
	for _, tz := range []string{"", "Mars/Olympus_Mons"} {
		store := newFakeStore()
		store.companies = []gen.ListReportCompaniesRow{company(1, tz)}
		store.recipients[1] = []string{"a@c.mx"}
		store.fuel[1] = []gen.ReportFuelByAssetRow{fuelRow("U")}
		sender := &fakeMail{}

		if err := newTestRunner(t, store, sender).Tick(context.Background()); err != nil {
			t.Fatal(err)
		}
		if len(sender.sent) != 1 {
			t.Fatalf("timezone %q: emails sent = %d, want 1", tz, len(sender.sent))
		}
	}
}

func TestForceSendsAgain(t *testing.T) {
	store := newFakeStore()
	store.recipients[1] = []string{"a@c.mx"}
	store.fuel[1] = []gen.ReportFuelByAssetRow{fuelRow("U")}
	sender := &fakeMail{}
	r := newTestRunner(t, store, sender)

	c := reports.Company{ID: 1, Name: "C", Timezone: "America/Mexico_City", Currency: "MXN"}
	loc, _ := reports.LoadLocation(c.Timezone)
	p := reports.LatestDue(reports.FuelWeekly, monday(t), loc, 6)

	if s, _ := r.RunOne(context.Background(), c, reports.FuelWeekly, p, false); s != StatusSent {
		t.Fatalf("first run = %s", s)
	}
	if s, _ := r.RunOne(context.Background(), c, reports.FuelWeekly, p, false); s != StatusNotClaimed {
		t.Fatalf("second run = %s, want not claimed", s)
	}
	if s, _ := r.RunOne(context.Background(), c, reports.FuelWeekly, p, true); s != StatusSent {
		t.Fatalf("forced run = %s", s)
	}
	if len(sender.sent) != 2 {
		t.Fatalf("emails sent = %d, want 2", len(sender.sent))
	}
}

// The controller ruling behind FinishReportRun's attempts predicate: a run
// taken over by a forced claim before the original owner finishes must not
// have its result overwritten by that stalled owner. This is a rule the
// generated SQL enforces (and RunOne relies on, logging a Warn and leaving
// the computed status/error untouched when FinishReportRun affects no rows);
// simulating the interleaving through two RunOne calls isn't expressible
// with the given API, since a single goroutine's RunOne claims and finishes
// in one call with no way to suspend it mid-flight. So this test drives the
// fake directly, the same way the SQL would be exercised: claim, force-claim
// (taking over the same run row and bumping its attempts), then have the
// superseded owner try to finish with its now-stale attempts value.
func TestFinishReportRunRejectsASupersededOwner(t *testing.T) {
	store := newFakeStore()
	arg := gen.ClaimReportRunParams{
		CompanyID:   1,
		ReportKind:  string(reports.FuelWeekly),
		PeriodStart: time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
	}

	first, err := store.ClaimReportRun(context.Background(), arg)
	if err != nil {
		t.Fatal(err)
	}
	if first.Attempts != 1 {
		t.Fatalf("first claim attempts = %d, want 1", first.Attempts)
	}

	forced := arg
	forced.Force = true
	second, err := store.ClaimReportRun(context.Background(), forced)
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID || second.Attempts != 2 {
		t.Fatalf("forced claim = %+v, want the same id with attempts 2", second)
	}

	// The new owner (the takeover) finishes first, and is recorded.
	rows, err := store.FinishReportRun(context.Background(), gen.FinishReportRunParams{
		ID: second.ID, Attempts: second.Attempts, Status: StatusSent, Recipients: 2,
	})
	if err != nil || rows != 1 {
		t.Fatalf("takeover finish: rows=%d err=%v, want 1, nil", rows, err)
	}

	// The stalled original owner finishes late, with its now-stale attempts.
	rows, err = store.FinishReportRun(context.Background(), gen.FinishReportRunParams{
		ID: first.ID, Attempts: first.Attempts, Status: StatusFailed, Recipients: 0,
	})
	if err != nil || rows != 0 {
		t.Fatalf("stalled finish: rows=%d err=%v, want 0, nil", rows, err)
	}

	if got := store.finished[first.ID]; got.Status != StatusSent {
		t.Fatalf("stored result = %q, want the takeover's %q left untouched", got.Status, StatusSent)
	}
}

func TestStartRejectsABadExpression(t *testing.T) {
	if _, err := start(func(context.Context) error { return nil }, "cada lunes", quiet, time.Second); err == nil {
		t.Fatal("a bad cron expression was accepted")
	}
}

// stop must wait for the tick in flight, so the pool is not closed under it.
func TestStopWaitsForTheTickInFlight(t *testing.T) {
	started := make(chan struct{})
	var finished atomic.Bool
	var once sync.Once

	stop, err := start(func(context.Context) error {
		once.Do(func() { close(started) })
		time.Sleep(300 * time.Millisecond)
		finished.Store(true)
		return nil
	}, "@every 1s", quiet, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("the tick never ran")
	}
	stop()
	if !finished.Load() {
		t.Fatal("stop returned while the tick was still running")
	}
}

// A tick that ignores the drain window is cancelled rather than waited on forever.
func TestStopCancelsATickThatOverstays(t *testing.T) {
	started := make(chan struct{})
	var once sync.Once

	stop, err := start(func(ctx context.Context) error {
		once.Do(func() { close(started) })
		<-ctx.Done()
		return ctx.Err()
	}, "@every 1s", quiet, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	<-started

	done := make(chan struct{})
	go func() { stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("stop did not return")
	}
}
