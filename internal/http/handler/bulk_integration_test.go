//go:build integration

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"

	"fleet/internal/db/gen"
	"fleet/internal/platform/dbctx"
	"fleet/internal/platform/sheet"
)

// Queries built on dbctx.DB join the transaction the context carries, and
// dbctx.Begin inside it is a savepoint that can roll back alone.
func TestContextTransactionJoinsQueriesAndBegin(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	q := gen.New(dbctx.New(pool))

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	txCtx := dbctx.WithTx(ctx, tx)

	if _, err := q.CreateAccount(txCtx, "outer write"); err != nil {
		t.Fatalf("write through the context transaction: %v", err)
	}

	inner, err := dbctx.Begin(txCtx, pool)
	if err != nil {
		t.Fatalf("savepoint: %v", err)
	}
	if _, err := gen.New(inner).CreateAccount(txCtx, "inner write"); err != nil {
		t.Fatalf("write inside the savepoint: %v", err)
	}
	if err := inner.Rollback(txCtx); err != nil {
		t.Fatalf("roll back the savepoint: %v", err)
	}

	count := func(ctx context.Context, db gen.DBTX) int {
		var n int
		if err := db.QueryRow(ctx,
			"SELECT count(*) FROM account WHERE name IN ('outer write','inner write')").Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if n := count(txCtx, tx); n != 1 {
		t.Fatalf("inside the transaction: %d rows, want 1 (the savepoint's write undone, the outer kept)", n)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if n := count(ctx, pool); n != 0 {
		t.Fatalf("after rolling back: %d rows, want 0", n)
	}
}

// importFixture is an owner signed in to a fresh company.
type importFixture struct {
	srv     *httptest.Server
	pool    *pgxpool.Pool
	token   string
	company int64
}

func newImportFixture(t *testing.T) importFixture {
	t.Helper()
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	srv := httptest.NewServer(newIntegrationRouter(pool))
	t.Cleanup(srv.Close)
	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	_, _, _, ownerToken := provisionAccountOwner(t, srv.URL, platformToken)
	company := createCompany(t, srv.URL, ownerToken, "Importadora", fmt.Sprintf("TAX-IMP-%d", time.Now().UnixNano()))
	return importFixture{srv: srv, pool: pool, token: switchCompany(t, srv.URL, ownerToken, company), company: company}
}

// upload posts rows as an .xlsx to path (e.g. "/vendors/import?dry_run=true").
func (f importFixture) upload(t *testing.T, path string, rows [][]any, want int) []byte {
	t.Helper()
	var file bytes.Buffer
	if err := sheet.WriteXLSX(&file, []sheet.Sheet{{Name: "Datos", Rows: rows}}); err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", "datos.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(file.Bytes()); err != nil {
		t.Fatal(err)
	}
	mw.Close()
	req, _ := http.NewRequest(http.MethodPost, f.srv.URL+"/api/v1"+path, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+f.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != want {
		t.Fatalf("POST %s: got %d, want %d: %s", path, resp.StatusCode, want, out)
	}
	return out
}

func (f importFixture) total(t *testing.T, path string) int64 {
	t.Helper()
	var page struct {
		Total int64 `json:"total"`
	}
	getJSON(t, f.srv.URL+"/api/v1"+path, f.token, http.StatusOK, &page)
	return page.Total
}

func TestImportCreatesEveryRow(t *testing.T) {
	f := newImportFixture(t)
	body := f.upload(t, "/vendors/import", [][]any{
		{"name", "city", "is_fuel_vendor"},
		{"Llantera Sur", "Tijuana", "sí"},
		{"Diesel Norte", "Ensenada", "no"},
		{"Taller Centro", "", ""},
	}, http.StatusCreated)
	var report struct{ Rows, Created int }
	if err := json.Unmarshal(body, &report); err != nil || report.Rows != 3 || report.Created != 3 {
		t.Fatalf("report %s (%v)", body, err)
	}
	if got := f.total(t, "/vendors"); got != 3 {
		t.Fatalf("vendors after import: %d, want 3", got)
	}
}

// One bad row, found before the database or by it, means nothing is imported.
func TestImportIsAllOrNothing(t *testing.T) {
	f := newImportFixture(t)

	body := f.upload(t, "/vendors/import", [][]any{
		{"name", "is_fuel_vendor"},
		{"Uno", "sí"},
		{"Dos", "quizá"},
	}, http.StatusUnprocessableEntity)
	if !strings.Contains(body2str(body), `"row":3`) || !strings.Contains(body2str(body), `"column":"is_fuel_vendor"`) {
		t.Fatalf("expected row 3 / is_fuel_vendor in %s", body)
	}
	if got := f.total(t, "/vendors"); got != 0 {
		t.Fatalf("vendors after a failed import: %d, want 0", got)
	}

	// A duplicate part number fails in the database, on row 3, after row 2
	// was already inserted: the whole file must still roll back.
	body = f.upload(t, "/parts/import", [][]any{
		{"part_number", "description"},
		{"FIL-001", "Filtro de aceite"},
		{"FIL-001", "Duplicado"},
	}, http.StatusUnprocessableEntity)
	if !strings.Contains(body2str(body), `"row":3`) {
		t.Fatalf("expected the duplicate on row 3 in %s", body)
	}
	if got := f.total(t, "/parts"); got != 0 {
		t.Fatalf("parts after a failed import: %d, want 0", got)
	}
}

func TestImportDryRunWritesNothing(t *testing.T) {
	f := newImportFixture(t)
	body := f.upload(t, "/vendors/import?dry_run=true", [][]any{{"name"}, {"Uno"}, {"Dos"}}, http.StatusOK)
	if !strings.Contains(body2str(body), `"created":2`) || !strings.Contains(body2str(body), `"dry_run":true`) {
		t.Fatalf("dry-run report %s", body)
	}
	if got := f.total(t, "/vendors"); got != 0 {
		t.Fatalf("vendors after a dry run: %d, want 0", got)
	}
}

func TestImportResolvesReferencesByName(t *testing.T) {
	f := newImportFixture(t)
	var category struct {
		ID int64 `json:"id"`
	}
	postJSON(t, f.srv.URL+"/api/v1/part-categories", f.token, map[string]any{"name": "Filtros"}, http.StatusCreated, &category)

	f.upload(t, "/parts/import", [][]any{
		{"part_number", "part_category_id"},
		{"FIL-002", "  filtros "},
	}, http.StatusCreated)
	var parts struct {
		Data []struct {
			PartCategoryID *int64 `json:"part_category_id"`
		} `json:"data"`
	}
	getJSON(t, f.srv.URL+"/api/v1/parts", f.token, http.StatusOK, &parts)
	if len(parts.Data) != 1 || parts.Data[0].PartCategoryID == nil || *parts.Data[0].PartCategoryID != category.ID {
		t.Fatalf("part category not resolved by name: %+v", parts.Data)
	}

	body := f.upload(t, "/parts/import", [][]any{{"part_number", "part_category_id"}, {"FIL-003", "Frenos"}}, http.StatusUnprocessableEntity)
	if !strings.Contains(body2str(body), "no ") || !strings.Contains(body2str(body), "Frenos") {
		t.Fatalf("unknown name: %s", body)
	}
}

// Assets carry the vehicle subtype as vehicle.* columns in the same row.
func TestImportAssetsWithVehicleColumns(t *testing.T) {
	f := newImportFixture(t)
	f.upload(t, "/assets/import", [][]any{
		{"name", "vin_sn", "make", "vehicle.engine_serial"},
		{"Unidad 101", "VIN-101", "Freightliner", "ENG-101"},
	}, http.StatusCreated)
	var assets struct {
		Data []struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	getJSON(t, f.srv.URL+"/api/v1/assets", f.token, http.StatusOK, &assets)
	if len(assets.Data) != 1 {
		t.Fatalf("assets: %+v", assets.Data)
	}
	var vehicle struct {
		EngineSerial string `json:"engine_serial"`
	}
	getJSON(t, fmt.Sprintf("%s/api/v1/assets/%d/vehicle", f.srv.URL, assets.Data[0].ID), f.token, http.StatusOK, &vehicle)
	if vehicle.EngineSerial != "ENG-101" {
		t.Fatalf("vehicle subtype: %+v", vehicle)
	}
}

func TestImportCustomFieldColumns(t *testing.T) {
	f := newImportFixture(t)
	postJSON(t, f.srv.URL+"/api/v1/custom-field-definitions", f.token, map[string]any{
		"resource": "vendors", "key": "cost_centre", "label": "Centro de costo", "field_type": "text", "required": true,
	}, http.StatusCreated, nil)

	body := f.upload(t, "/vendors/import", [][]any{{"name", "cf.cost_centre"}, {"Sin centro", ""}}, http.StatusUnprocessableEntity)
	if !strings.Contains(body2str(body), "custom_fields.cost_centre") {
		t.Fatalf("missing required custom field: %s", body)
	}
	f.upload(t, "/vendors/import", [][]any{{"name", "cf.cost_centre"}, {"Con centro", "A12"}}, http.StatusCreated)
}

// A child section names its parent by name in the parent column.
//
// employee_id and vendor_id are NOT NULL foreign keys on fuel_entry (see
// 000001_init.up.sql), so this needs real rows to point at even though
// neither column is required by CreateFuelEntryRequest's own validation:
// omitting them would send id 0 and fail with a foreign-key violation rather
// than exercising the by-name asset resolution this test is about.
func TestImportNestedFuelEntriesByAssetName(t *testing.T) {
	f := newImportFixture(t)
	postJSON(t, f.srv.URL+"/api/v1/assets", f.token, map[string]any{"name": "Unidad 7", "vin_sn": "VIN-7"}, http.StatusCreated, nil)
	employeeID := createDirectorCandidate(t, f.srv.URL, f.token, "fuel")
	var vendor struct {
		ID int64 `json:"id"`
	}
	postJSON(t, f.srv.URL+"/api/v1/vendors", f.token, map[string]any{"name": "Gasolinera Uno"}, http.StatusCreated, &vendor)

	f.upload(t, "/fuel-entries/import", [][]any{
		{"asset_id", "date", "quantity", "unit_cost", "total_cost", "odometer", "employee_id", "vendor_id"},
		{"unidad 7", "2026-09-01", "120", "24.5", "2940", "150000", fmt.Sprint(employeeID), fmt.Sprint(vendor.ID)},
		{"Unidad 7", "02/09/2026", "110", "24.7", "2717", "150800", fmt.Sprint(employeeID), fmt.Sprint(vendor.ID)},
	}, http.StatusCreated)
	if got := f.total(t, "/fuel-entries"); got != 2 {
		t.Fatalf("fuel entries: %d, want 2", got)
	}
}

func TestImportTemplate(t *testing.T) {
	f := newImportFixture(t)
	postJSON(t, f.srv.URL+"/api/v1/part-categories", f.token, map[string]any{"name": "Filtros"}, http.StatusCreated, nil)

	req, _ := http.NewRequest(http.MethodGet, f.srv.URL+"/api/v1/parts/import/template", nil)
	req.Header.Set("Authorization", "Bearer "+f.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !strings.Contains(resp.Header.Get("Content-Disposition"), "parts-plantilla.xlsx") {
		t.Fatalf("template: %d %v", resp.StatusCode, resp.Header)
	}
	file, _ := io.ReadAll(resp.Body)
	rows, err := sheet.Read(bytes.NewReader(file), "parts-plantilla.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	header := strings.Join(rows[0], ",")
	for _, want := range []string{"part_number", "part_category_id", "unit_cost"} {
		if !strings.Contains(header, want) {
			t.Fatalf("template header %q lacks %s", header, want)
		}
	}
	// The workbook is a zip of DEFLATE-compressed XML parts, so a plain byte
	// search for "Filtros" in the raw file would not find it even if present;
	// open it and read the Catálogos sheet's actual cells instead.
	wb, err := excelize.OpenReader(bytes.NewReader(file))
	if err != nil {
		t.Fatal(err)
	}
	defer wb.Close()
	catalog, err := wb.GetRows(catalogSheetName)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range catalog {
		if slices.Contains(row, "Filtros") {
			found = true
		}
	}
	if !found {
		t.Fatalf("the %s sheet does not list the company's part categories: %v", catalogSheetName, catalog)
	}
}

// catalogSheetName mirrors bulk.Template's own sheet name (unexported there);
// duplicated here rather than exported for one test to read.
const catalogSheetName = "Catálogos"

// Import sits on the section's own route group: no module permission, no import.
func TestImportNeedsTheSectionPermission(t *testing.T) {
	f := newImportFixture(t)
	memberID := createDirectorCandidate(t, f.srv.URL, f.token, "noimport")
	f.token = setPasswordAndLogIn(t, f.srv.URL, f.token, memberID, employeeEmail(t, f.srv.URL, f.token, memberID))
	f.upload(t, "/vendors/import", [][]any{{"name"}, {"Uno"}}, http.StatusForbidden)
}

func TestImportRowLimit(t *testing.T) {
	t.Setenv("IMPORT_MAX_ROWS", "2")
	f := newImportFixture(t)
	f.upload(t, "/vendors/import", [][]any{{"name"}, {"Uno"}, {"Dos"}, {"Tres"}}, http.StatusBadRequest)
}

// Multi-tenancy is a security property here: a reference lookup pages through
// the referenced section's own store (bulkLookups' doc comment says so), which
// scopes every query to middleware.CompanyFromContext. This proves that scope
// actually holds for an import — by name and by explicit id — and that the
// created rows themselves land only in the caller's own company.
func TestImportReferencesAreScopedToTheCallersCompany(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	srv := httptest.NewServer(newIntegrationRouter(pool))
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	_, _, _, ownerToken := provisionAccountOwner(t, srv.URL, platformToken)

	ts := time.Now().UnixNano()
	companyA := createCompany(t, srv.URL, ownerToken, "Company A", fmt.Sprintf("TAX-XA-%d", ts))
	companyB := createCompany(t, srv.URL, ownerToken, "Company B", fmt.Sprintf("TAX-XB-%d", ts))
	tokenA := switchCompany(t, srv.URL, ownerToken, companyA)
	tokenB := switchCompany(t, srv.URL, ownerToken, companyB)

	// A part category that exists only in B.
	var categoryB struct {
		ID int64 `json:"id"`
	}
	postJSON(t, srv.URL+"/api/v1/part-categories", tokenB, map[string]any{"name": "SoloB"}, http.StatusCreated, &categoryB)

	f := importFixture{srv: srv, pool: pool, token: tokenA, company: companyA}

	// By name: a name that exists only in B is unresolved from A.
	body := f.upload(t, "/parts/import", [][]any{{"part_number", "part_category_id"}, {"XA-1", "SoloB"}}, http.StatusUnprocessableEntity)
	if !strings.Contains(body2str(body), "SoloB") {
		t.Fatalf("expected the cross-company name to be reported as unresolved, got %s", body)
	}

	// By explicit id: B's numeric category id is refused from A rather than
	// silently linking the new part to another tenant's row.
	body = f.upload(t, "/parts/import", [][]any{{"part_number", "part_category_id"}, {"XA-2", fmt.Sprint(categoryB.ID)}}, http.StatusUnprocessableEntity)
	if !strings.Contains(body2str(body), fmt.Sprint(categoryB.ID)) {
		t.Fatalf("expected company B's id %d to be refused from company A, got %s", categoryB.ID, body)
	}
	if got := f.total(t, "/parts"); got != 0 {
		t.Fatalf("company A parts after two refused imports: %d, want 0", got)
	}

	// A successful import lands only in the caller's own company.
	f.upload(t, "/parts/import", [][]any{{"part_number"}, {"XA-3"}}, http.StatusCreated)
	if got := f.total(t, "/parts"); got != 1 {
		t.Fatalf("company A parts: %d, want 1", got)
	}
	var partsB struct {
		Total int64 `json:"total"`
	}
	getJSON(t, srv.URL+"/api/v1/parts", tokenB, http.StatusOK, &partsB)
	if partsB.Total != 0 {
		t.Fatalf("company B parts: %d, want 0 (the import leaked across companies)", partsB.Total)
	}
}

func body2str(b []byte) string { return string(b) }
