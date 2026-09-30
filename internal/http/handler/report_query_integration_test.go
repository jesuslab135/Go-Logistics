//go:build integration

package handler

// The scheduled reports' SQL against a real Postgres: whether the two
// aggregates select the right rows, and whether the claim on report_run has
// exactly one winner. The arithmetic and the rendering are unit-tested in
// internal/platform/reports and internal/platform/pdf.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"fleet/internal/db/gen"
	"fleet/internal/platform/dbctx"
	"fleet/internal/platform/reports"
)

// reportFixture is an owner signed in to a fresh company, plus the pool, so a
// test can set columns the API computes or does not expose.
type reportFixture struct {
	srv     string
	token   string // the owner, scoped to company
	owner   string // the owner, scoped to no company
	company int64
	pool    *pgxpool.Pool
	q       *gen.Queries
	loc     *time.Location
}

func newReportFixture(t *testing.T) reportFixture {
	t.Helper()
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	srv := httptest.NewServer(newIntegrationRouter(pool))
	t.Cleanup(srv.Close)

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	_, _, _, owner := provisionAccountOwner(t, srv.URL, platformToken)
	company := createCompany(t, srv.URL, owner, "Transportes Reporte", fmt.Sprintf("TAX-REP-%d", time.Now().UnixNano()))

	loc, err := time.LoadLocation("America/Mexico_City")
	if err != nil {
		t.Fatal(err)
	}
	return reportFixture{
		srv:     srv.URL,
		token:   switchCompany(t, srv.URL, owner, company),
		owner:   owner,
		company: company,
		pool:    pool,
		q:       gen.New(dbctx.New(pool)),
		loc:     loc,
	}
}

// limitedToken logs in as a member of the company who holds no role, and so
// no module at all.
func (f reportFixture) limitedToken(t *testing.T) string {
	t.Helper()
	id := createDirectorCandidate(t, f.srv, f.token, "sin-reportes")
	return setPasswordAndLogIn(t, f.srv, f.token, id, employeeEmail(t, f.srv, f.token, id))
}

func (f reportFixture) exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := f.pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

func (f reportFixture) asset(t *testing.T, token, name string) int64 {
	t.Helper()
	var out struct {
		ID int64 `json:"id"`
	}
	postJSON(t, f.srv+"/api/v1/assets", token, map[string]any{"name": name, "vin_sn": "VIN-" + name}, http.StatusCreated, &out)
	return out.ID
}

// fuel inserts a fuel entry directly. The API derives miles_traveled and
// fuel_efficiency from the series; the report only reads them, so the test
// states them.
func (f reportFixture) fuel(t *testing.T, assetID, employeeID, vendorID int64, at time.Time, quantity, cost, miles string, efficiency any) {
	t.Helper()
	f.exec(t, `
		INSERT INTO fuel_entry (
			asset_id, employee_id, date, fuel_type, quantity, unit_cost, total_cost, odometer,
			vendor_id, full_tank, miles_traveled, fuel_efficiency, state, reference, external_id, updated_at
		) VALUES ($1,$2,$3,'diesel',$4,1,$5,0,$6,true,$7,$8,'','','',now())`,
		assetID, employeeID, at, quantity, cost, vendorID, miles, efficiency)
}

func TestReportFuelQuery(t *testing.T) {
	f := newReportFixture(t)
	ctx := context.Background()

	asset := f.asset(t, f.token, "Unidad 7")
	employee := createDirectorCandidate(t, f.srv, f.token, "fuel-report")
	var vendor struct {
		ID int64 `json:"id"`
	}
	postJSON(t, f.srv+"/api/v1/vendors", f.token, map[string]any{"name": "Gasolinera"}, http.StatusCreated, &vendor)

	// Week of Monday 2026-09-21, Mexico City.
	week := reports.WeekOf(time.Date(2026, 9, 23, 12, 0, 0, 0, f.loc), f.loc)

	f.fuel(t, asset, employee, vendor.ID, week.Start, "100", "2500", "0", nil)                        // first instant: in
	f.fuel(t, asset, employee, vendor.ID, week.Start.Add(48*time.Hour), "120", "3000", "800", "6.67") // in, with efficiency
	f.fuel(t, asset, employee, vendor.ID, week.End.Add(-time.Second), "80", "2000", "500", nil)       // last second: in
	f.fuel(t, asset, employee, vendor.ID, week.Start.Add(-time.Second), "999", "9999", "999", "1")    // the second before: out
	f.fuel(t, asset, employee, vendor.ID, week.End, "999", "9999", "999", "1")                        // the end itself: out

	// Another tenant's fuel, in the same week, must not appear.
	other := createCompany(t, f.srv, f.owner, "Otra Empresa", fmt.Sprintf("TAX-OTRA-%d", time.Now().UnixNano()))
	otherToken := switchCompany(t, f.srv, f.owner, other)
	otherAsset := f.asset(t, otherToken, "Ajena")
	f.fuel(t, otherAsset, employee, vendor.ID, week.Start.Add(time.Hour), "777", "7777", "777", "1")

	rows, err := f.q.ReportFuelByAsset(ctx, gen.ReportFuelByAssetParams{
		CompanyID: f.company, PeriodStart: week.Start, PeriodEnd: week.End, Timezone: f.loc.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1: %+v", len(rows), rows)
	}
	r := rows[0]
	if r.AssetName != "Unidad 7" || r.Fills != 3 {
		t.Fatalf("row = %+v", r)
	}
	wantDec(t, "volume", r.Volume, "300")
	wantDec(t, "cost", r.Cost, "7500")
	wantDec(t, "distance", r.Distance, "1300")
	// Only the entry with a derived efficiency counts towards the ratio.
	wantDec(t, "eff_distance", r.EffDistance, "800")
	wantDec(t, "eff_volume", r.EffVolume, "120")
}

func wantDec(t *testing.T, name string, got decimal.Decimal, want string) {
	t.Helper()
	if !got.Equal(decimal.RequireFromString(want)) {
		t.Fatalf("%s = %s, want %s", name, got, want)
	}
}

// statuses returns the id of the company's status that marks a work
// order as completed, and of one that does not.
func (f reportFixture) statuses(t *testing.T) (completed, open int64) {
	t.Helper()
	ctx := context.Background()
	if err := f.pool.QueryRow(ctx,
		`SELECT id FROM work_order_status WHERE company_id = $1 AND marks_as_completed ORDER BY id LIMIT 1`, f.company,
	).Scan(&completed); err != nil {
		t.Fatalf("completed status: %v", err)
	}
	if err := f.pool.QueryRow(ctx,
		`SELECT id FROM work_order_status WHERE company_id = $1 AND NOT marks_as_completed ORDER BY id LIMIT 1`, f.company,
	).Scan(&open); err != nil {
		t.Fatalf("open status: %v", err)
	}
	return completed, open
}

func (f reportFixture) workOrder(t *testing.T, asset, status int64, completedAt *time.Time, parts, labor, total string) int64 {
	t.Helper()
	var out struct {
		ID int64 `json:"id"`
	}
	// Noon UTC on 2026-08-15 is 06:00 in Mexico City: solidly inside August
	// there, unlike a UTC midnight boundary which can fall on the wrong side
	// of the local date.
	postJSON(t, f.srv+"/api/v1/work-orders", f.token, map[string]any{
		"asset_id": asset, "status_id": status, "issued_at": time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC),
		"completed_at": completedAt, "number": fmt.Sprintf("WO-%d", time.Now().UnixNano()),
	}, http.StatusCreated, &out)
	// The API computes money from line items; the report reads the stored
	// columns, so the test writes them.
	f.exec(t, `UPDATE work_order SET parts_subtotal = $2, labor_subtotal = $3, total_amount = $4 WHERE id = $1`,
		out.ID, parts, labor, total)
	return out.ID
}

func (f reportFixture) serviceEntry(t *testing.T, asset int64, workOrder *int64, completedAt *time.Time, parts, labor, total string) int64 {
	t.Helper()
	var out struct {
		ID int64 `json:"id"`
	}
	postJSON(t, f.srv+"/api/v1/service-entries", f.token, map[string]any{
		"asset_id": asset, "status": "COMPLETED", "work_order_id": workOrder, "completed_at": completedAt,
	}, http.StatusCreated, &out)
	f.exec(t, `UPDATE service_entry SET parts_subtotal = $2, labor_subtotal = $3, total_amount = $4 WHERE id = $1`,
		out.ID, parts, labor, total)
	return out.ID
}

func TestReportMaintenanceQuery(t *testing.T) {
	f := newReportFixture(t)
	ctx := context.Background()

	asset := f.asset(t, f.token, "Unidad 7")
	completed, open := f.statuses(t)

	month := reports.MonthOf(time.Date(2026, 8, 15, 0, 0, 0, 0, f.loc), f.loc)
	inMonth := month.Start.Add(10 * 24 * time.Hour)
	before := month.Start.Add(-time.Hour)

	// 1. A completed work order on its own: counted.
	f.workOrder(t, asset, completed, &inMonth, "100", "50", "150")
	// 2. A completed work order with a linked service entry: one job, and the
	//    service entry's figures are the ones counted.
	linked := f.workOrder(t, asset, completed, &inMonth, "9000", "9000", "18000")
	f.serviceEntry(t, asset, &linked, &inMonth, "200", "100", "300")
	// 3. A service entry with a manual override: the override is the total.
	overridden := f.serviceEntry(t, asset, nil, &inMonth, "10", "10", "20")
	f.exec(t, `UPDATE service_entry SET total_override = 1000 WHERE id = $1`, overridden)
	// 4. A work order that is not completed: not counted.
	f.workOrder(t, asset, open, &inMonth, "7000", "7000", "14000")
	// 5. A completed work order from the month before: not counted.
	f.workOrder(t, asset, completed, &before, "5000", "5000", "10000")

	rows, err := f.q.ReportMaintenanceByAsset(ctx, gen.ReportMaintenanceByAssetParams{
		CompanyID: f.company, PeriodStart: month.Start, PeriodEnd: month.End, Timezone: f.loc.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1: %+v", len(rows), rows)
	}
	r := rows[0]
	if r.Jobs != 3 {
		t.Fatalf("jobs = %d, want 3 (the linked pair is one job)", r.Jobs)
	}
	wantDec(t, "parts", r.Parts, "310")
	wantDec(t, "labor", r.Labor, "160")
	wantDec(t, "total", r.Total, "1450") // 150 + 300 + 1000
	if !r.HasOverride {
		t.Fatal("has_override = false, want true")
	}
}

// A completed job with no completed_at is dated by a stable column instead,
// so it is reported exactly once rather than never, or twice if that column
// could drift. A work order falls back to issued_at (fixed at creation); a
// service entry falls back to created_at (also fixed at creation, here pinned
// explicitly so the test does not depend on when it runs).
func TestReportMaintenanceCountsJobsWithoutACompletionDate(t *testing.T) {
	f := newReportFixture(t)
	ctx := context.Background()
	asset := f.asset(t, f.token, "Unidad 7")
	completed, _ := f.statuses(t)

	// Dateless work order: falls back to issued_at, which the helper fixes at
	// 2026-08-15, well inside August in Mexico City local time.
	f.workOrder(t, asset, completed, nil, "100", "50", "150")

	// Dateless service entry: falls back to created_at. Pin it to October so
	// it lands in a month of its own, distinct from the work order's August.
	entry := f.serviceEntry(t, asset, nil, nil, "200", "100", "300")
	f.exec(t, `UPDATE service_entry SET created_at = $2 WHERE id = $1`,
		entry, time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC))

	august := reports.MonthOf(time.Date(2026, 8, 15, 0, 0, 0, 0, f.loc), f.loc)
	september := reports.MonthOf(time.Date(2026, 9, 15, 0, 0, 0, 0, f.loc), f.loc)
	october := reports.MonthOf(time.Date(2026, 10, 15, 0, 0, 0, 0, f.loc), f.loc)

	// Included: the work order belongs to August (its issued_at).
	augRows, err := f.q.ReportMaintenanceByAsset(ctx, gen.ReportMaintenanceByAssetParams{
		CompanyID: f.company, PeriodStart: august.Start, PeriodEnd: august.End, Timezone: f.loc.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(augRows) != 1 || augRows[0].Jobs != 1 {
		t.Fatalf("august rows = %+v, want one row with 1 job (the work order)", augRows)
	}
	wantDec(t, "august total", augRows[0].Total, "150")

	// Excluded: the work order must not also appear in September, which
	// would happen if it were dated by a column that moves after creation.
	sepRows, err := f.q.ReportMaintenanceByAsset(ctx, gen.ReportMaintenanceByAssetParams{
		CompanyID: f.company, PeriodStart: september.Start, PeriodEnd: september.End, Timezone: f.loc.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sepRows) != 0 {
		t.Fatalf("september rows = %+v, want none (the work order belongs to august only)", sepRows)
	}

	// Included: the service entry belongs to October (its pinned created_at).
	octRows, err := f.q.ReportMaintenanceByAsset(ctx, gen.ReportMaintenanceByAssetParams{
		CompanyID: f.company, PeriodStart: october.Start, PeriodEnd: october.End, Timezone: f.loc.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(octRows) != 1 || octRows[0].Jobs != 1 {
		t.Fatalf("october rows = %+v, want one row with 1 job (the service entry)", octRows)
	}
	// Excluded: the service entry (300) must not also be folded into August's
	// total (150) — the earlier august assertions already prove this, since
	// augRows had exactly 1 job and a total of 150, not 2 jobs and 450.
	wantDec(t, "october total", octRows[0].Total, "300")
}

// A date picked without a time reaches the API as that day's UTC midnight
// (what JavaScript's toISOString and the bulk importer both produce). In
// Mexico City that instant is 18:00 the evening before, so reading it as a
// moment would move a Monday fill into the previous week and a 1st-of-month
// job into the previous month. The reports read an exact UTC midnight as the
// calendar date it was written as, and any other instant as the local day it
// falls on.
func TestReportsReadUTCMidnightAsACalendarDate(t *testing.T) {
	f := newReportFixture(t)
	ctx := context.Background()
	tz := f.loc.String()

	asset := f.asset(t, f.token, "Unidad 7")
	employee := createDirectorCandidate(t, f.srv, f.token, "bare-dates")
	var vendor struct {
		ID int64 `json:"id"`
	}
	postJSON(t, f.srv+"/api/v1/vendors", f.token, map[string]any{"name": "Gasolinera"}, http.StatusCreated, &vendor)

	// Week of Monday 2026-09-21 to Sunday 2026-09-27, Mexico City.
	week := reports.WeekOf(time.Date(2026, 9, 23, 12, 0, 0, 0, f.loc), f.loc)
	f.fuel(t, asset, employee, vendor.ID, time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC), "10", "100", "0", nil)   // Monday, date only: in
	f.fuel(t, asset, employee, vendor.ID, time.Date(2026, 9, 27, 23, 30, 0, 0, f.loc), "20", "200", "0", nil)    // Sunday night, a real moment: in
	f.fuel(t, asset, employee, vendor.ID, time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC), "999", "9999", "0", nil) // next Monday, date only: out
	f.fuel(t, asset, employee, vendor.ID, time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC), "999", "9999", "0", nil) // previous Sunday, date only: out

	fuel, err := f.q.ReportFuelByAsset(ctx, gen.ReportFuelByAssetParams{
		CompanyID: f.company, PeriodStart: week.Start, PeriodEnd: week.End, Timezone: tz,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fuel) != 1 || fuel[0].Fills != 2 {
		t.Fatalf("fuel rows = %+v, want one row with 2 fills", fuel)
	}
	wantDec(t, "fuel cost", fuel[0].Cost, "300")

	completed, _ := f.statuses(t)
	month := reports.MonthOf(time.Date(2026, 8, 15, 0, 0, 0, 0, f.loc), f.loc)
	first := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)       // 1 August, date only: in
	lastNight := time.Date(2026, 8, 31, 23, 0, 0, 0, f.loc)    // 31 August 23:00 local: in
	nextFirst := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)   // 1 September, date only: out
	lastOfJuly := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC) // 31 July, date only: out
	f.workOrder(t, asset, completed, &first, "0", "0", "100")
	f.serviceEntry(t, asset, nil, &lastNight, "0", "0", "200")
	f.workOrder(t, asset, completed, &nextFirst, "0", "0", "9999")
	f.serviceEntry(t, asset, nil, &lastOfJuly, "0", "0", "9999")

	maint, err := f.q.ReportMaintenanceByAsset(ctx, gen.ReportMaintenanceByAssetParams{
		CompanyID: f.company, PeriodStart: month.Start, PeriodEnd: month.End, Timezone: tz,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(maint) != 1 || maint[0].Jobs != 2 {
		t.Fatalf("maintenance rows = %+v, want one row with 2 jobs", maint)
	}
	wantDec(t, "maintenance total", maint[0].Total, "300")
}

// A stalled or superseded owner finishing late must not overwrite the result
// of whoever holds the run now: FinishReportRun matches on attempts, which a
// takeover always bumps, so a stale finish affects no row.
func TestFinishReportRunOnlyAffectsTheCurrentOwner(t *testing.T) {
	f := newReportFixture(t)
	ctx := context.Background()
	week := reports.WeekOf(time.Date(2026, 9, 23, 0, 0, 0, 0, f.loc), f.loc)

	first, err := f.q.ClaimReportRun(ctx, gen.ClaimReportRunParams{
		CompanyID: f.company, ReportKind: "fuel_weekly", PeriodStart: week.StartDate(),
		MaxAttempts: 3, StaleMinutes: 30,
	})
	if err != nil {
		t.Fatalf("first claim: %v", err)
	}

	// A forced takeover: the same row, attempts bumped.
	second, err := f.q.ClaimReportRun(ctx, gen.ClaimReportRunParams{
		CompanyID: f.company, ReportKind: "fuel_weekly", PeriodStart: week.StartDate(),
		Force: true, MaxAttempts: 3, StaleMinutes: 30,
	})
	if err != nil {
		t.Fatalf("second (forced) claim: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("second claim id = %d, want the same row %d", second.ID, first.ID)
	}
	if second.Attempts != first.Attempts+1 {
		t.Fatalf("second claim attempts = %d, want %d", second.Attempts, first.Attempts+1)
	}

	// The stalled first owner finishes late, carrying the attempts value it
	// claimed with. It must not touch the row the second owner now holds.
	n, err := f.q.FinishReportRun(ctx, gen.FinishReportRunParams{
		ID: first.ID, Attempts: first.Attempts, Status: "failed", Recipients: 0,
	})
	if err != nil {
		t.Fatalf("finish (stale owner): %v", err)
	}
	if n != 0 {
		t.Fatalf("stale finish affected %d rows, want 0", n)
	}

	var status string
	if err := f.pool.QueryRow(ctx, `SELECT status FROM report_run WHERE id = $1`, first.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "running" {
		t.Fatalf("status after the stale finish = %q, want running (unchanged)", status)
	}

	// The current owner finishes normally, with its own attempts value.
	n, err = f.q.FinishReportRun(ctx, gen.FinishReportRunParams{
		ID: second.ID, Attempts: second.Attempts, Status: "sent", Recipients: 2,
	})
	if err != nil {
		t.Fatalf("finish (current owner): %v", err)
	}
	if n != 1 {
		t.Fatalf("current-owner finish affected %d rows, want 1", n)
	}

	if err := f.pool.QueryRow(ctx, `SELECT status FROM report_run WHERE id = $1`, second.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "sent" {
		t.Fatalf("status = %q, want sent", status)
	}
}

// The claim alone, without the advisory lock in front of it.
func TestConcurrentClaimsHaveOneWinner(t *testing.T) {
	f := newReportFixture(t)
	week := reports.WeekOf(time.Date(2026, 9, 23, 0, 0, 0, 0, f.loc), f.loc)

	var won atomic.Int64
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := f.q.ClaimReportRun(context.Background(), gen.ClaimReportRunParams{
				CompanyID: f.company, ReportKind: "fuel_weekly", PeriodStart: week.StartDate(),
				MaxAttempts: 3, StaleMinutes: 30,
			})
			switch {
			case err == nil:
				won.Add(1)
			case errors.Is(err, pgx.ErrNoRows):
			default:
				t.Errorf("claim: %v", err)
			}
		}()
	}
	wg.Wait()
	if got := won.Load(); got != 1 {
		t.Fatalf("claims won = %d, want 1", got)
	}
}
