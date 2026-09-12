package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"fleet/internal/http/dto"
	"fleet/internal/platform/bulk"
)

// A handler that calls pool.Begin opens a second transaction, outside the one a
// bulk import carries on the context, so its writes would survive a failed
// import. dbctx.Begin joins the import's transaction as a savepoint instead.
func TestHandlersDoNotBeginTransactionsDirectly(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(src, []byte("pool.Begin(")) {
			t.Errorf("%s calls pool.Begin directly; use dbctx.Begin(ctx, pool)", f)
		}
	}
}

// A handler that reaches for the pool directly bypasses the transaction a bulk
// import carries on the context: its writes would outlive a failed import and
// its reads would not see the import's own rows. dbctx.New(pool) routes both
// through the context's transaction when there is one.
//
// Two files are exempt and say so in their own comments: runList
// (listquery.go) and facetQuery (facets.go) are read-only and never run inside
// an import's transaction. Anything else is a real bug, which is the half of
// the dbctx invariant this guard adds on top of the Begin check above.
func TestHandlersUseDbctxRatherThanThePoolDirectly(t *testing.T) {
	exempt := map[string]bool{"listquery.go": true, "facets.go": true}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") || exempt[f] {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, call := range []string{"pool.Query(", "pool.QueryRow(", "pool.Exec("} {
			if bytes.Contains(src, []byte(call)) {
				t.Errorf("%s calls %s directly; use dbctx.New(pool) so it joins the request's transaction", f, call)
			}
		}
	}
}

// stubImporter is an Importer that would succeed if it ever ran. The
// concurrency test below asserts that a refused import never reaches it.
type stubImporter struct{ prepared bool }

func (s *stubImporter) Resource() string { return "vendors" }
func (s *stubImporter) Columns(context.Context) ([]bulk.Column, error) {
	return []bulk.Column{{Key: "name", Kind: bulk.KindString}}, nil
}
func (s *stubImporter) Refs() map[string]bulk.Lookup { return map[string]bulk.Lookup{} }
func (s *stubImporter) Prepare(map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
	s.prepared = true
	return func(context.Context) error { return nil }, nil
}

// An import that cannot get a slot must be refused straight away, not queued.
// bulk.Run holds a pooled connection for the entire import, so the point of
// the limit is to keep the pool from being drained; a caller parked behind a
// multi-minute import has already given up, and the server sets no write
// timeout that would ever break the wait.
func TestImportRefusesWhenAllSlotsAreBusy(t *testing.T) {
	h := &BulkHandler{maxBytes: 1 << 20, limits: bulk.Limits{}, slots: make(chan struct{}, 1)}
	h.slots <- struct{}{} // the only slot belongs to an import already running

	imp := &stubImporter{}
	router := gin.New()
	router.POST("/vendors/import", h.importRows(imp))

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", "datos.csv")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write([]byte("name\nUno\n")); err != nil {
		t.Fatal(err)
	}
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/vendors/import", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		router.ServeHTTP(rec, req)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the import blocked waiting for a slot; it must refuse immediately instead")
	}

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d: %s", rec.Code, http.StatusTooManyRequests, rec.Body)
	}
	if imp.prepared {
		t.Fatal("a refused import must not have started importing rows")
	}
	// The slot the running import holds is still its own.
	if len(h.slots) != 1 {
		t.Fatalf("slots in use = %d; want 1 (a refusal must not take or release one)", len(h.slots))
	}
}

// Every reference a section declares must be a column its create request
// actually has; a typo would otherwise silently disable the lookup.
func TestBulkRegistryReferencesExistingColumns(t *testing.T) {
	for _, e := range bulkImporters(Deps{}) {
		cols, err := e.imp.Columns(context.Background())
		if err != nil {
			// Sections with custom fields need a database; check their static columns.
			t.Logf("%s: %v (custom fields skipped)", e.imp.Resource(), err)
			continue
		}
		keys := map[string]bool{}
		for _, c := range cols {
			keys[c.Key] = true
		}
		for ref := range e.imp.Refs() {
			if !keys[ref] {
				t.Errorf("%s declares a reference on %q, which is not one of its columns", e.imp.Resource(), ref)
			}
		}
	}
}

func TestBulkRoutesAreRegistered(t *testing.T) {
	routes := map[string]bool{}
	for _, r := range newTestRouter(t).Routes() {
		routes[r.Method+" "+r.Path] = true
	}
	for _, e := range bulkImporters(Deps{}) {
		for _, want := range []string{"POST /api/v1" + e.path + "/import", "GET /api/v1" + e.path + "/import/template"} {
			if !routes[want] {
				t.Errorf("missing route %s", want)
			}
		}
	}
}

// gin's route table (what TestBulkRoutesAreRegistered reads) carries no
// middleware, so a section pointed at the wrong bulkGroups field — the wrong
// module gate — would still show up there as "registered" with nothing to
// catch it. This asserts each entry's declared group directly against the
// module router.go actually assigns that path, by giving every bulkGroups
// field a distinct, identifiable *gin.RouterGroup and checking e.group(bg)
// returns the one this table says it should.
func TestBulkEntriesUseTheirDeclaredGroup(t *testing.T) {
	engine := gin.New()
	bg := bulkGroups{
		member:       engine.Group("/g/member"),
		assets:       engine.Group("/g/assets"),
		tires:        engine.Group("/g/tires"),
		parts:        engine.Group("/g/parts"),
		inventory:    engine.Group("/g/inventory"),
		workOrders:   engine.Group("/g/work-orders"),
		issues:       engine.Group("/g/issues"),
		service:      engine.Group("/g/service"),
		fuel:         engine.Group("/g/fuel"),
		vendors:      engine.Group("/g/vendors"),
		warranties:   engine.Group("/g/warranties"),
		mileageGoals: engine.Group("/g/mileage-goals"),
		employees:    engine.Group("/g/employees"),
	}

	// Mirrors the group given to each entry in bulk_registry.go.
	want := map[string]gin.IRouter{
		"/assets":                       bg.assets,
		"/vendors":                      bg.vendors,
		"/locations":                    bg.member,
		"/part-categories":              bg.parts,
		"/part-manufacturers":           bg.parts,
		"/parts":                        bg.parts,
		"/measurement-units":            bg.inventory,
		"/part-locations":               bg.inventory,
		"/inventory-adjustment-reasons": bg.inventory,
		"/part-inventory":               bg.inventory,
		"/work-order-statuses":          bg.workOrders,
		"/work-orders":                  bg.workOrders,
		"/issues":                       bg.issues,
		"/issue-priorities":             bg.issues,
		"/faults":                       bg.issues,
		"/service-tasks":                bg.service,
		"/service-reminders":            bg.service,
		"/service-entries":              bg.service,
		"/tire-models":                  bg.tires,
		"/tires":                        bg.tires,
		"/tire-inspections":             bg.tires,
		"/fuel-types":                   bg.fuel,
		"/fuel-entries":                 bg.fuel,
		"/trailer-classifications":      bg.assets,
		"/warranties":                   bg.warranties,
		"/weekly-mileage-goals":         bg.mileageGoals,
		"/employees":                    bg.employees,
		"/groups":                       bg.employees,
		"/asset-types":                  bg.assets,
		"/asset-statuses":               bg.assets,
		"/catalog-options":              bg.assets,
		"/vehicle-makes":                bg.member,
		"/vehicle-models":               bg.member,
	}

	entries := bulkImporters(Deps{})
	if len(entries) != len(want) {
		t.Fatalf("bulkImporters has %d entries but this test's table has %d; keep them in sync", len(entries), len(want))
	}
	seen := map[string]bool{}
	for _, e := range entries {
		seen[e.path] = true
		wantGroup, ok := want[e.path]
		if !ok {
			t.Errorf("%s: no expected group declared in this test", e.path)
			continue
		}
		if got := e.group(bg); got != wantGroup {
			t.Errorf("%s: registered on the wrong module group", e.path)
		}
	}
	for path := range want {
		if !seen[path] {
			t.Errorf("%s: declared in this test's table but not in bulkImporters", path)
		}
	}
}

// Every export path must be a list route that exists, or the export would
// relay a 404 for every call.
func TestExportPathsAreListRoutes(t *testing.T) {
	routes := map[string]bool{}
	for _, r := range newTestRouter(t).Routes() {
		routes[r.Method+" "+r.Path] = true
	}
	for _, p := range exportPaths {
		if !routes["GET /api/v1"+p] {
			t.Errorf("export path %s has no GET list route", p)
		}
		if !routes["GET /api/v1"+p+"/export"] {
			t.Errorf("export route for %s is not registered", p)
		}
	}
}

// The documented report must have exactly the JSON shape the engine returns.
func TestImportReportDocMatchesEngine(t *testing.T) {
	keys := func(v any) []string {
		b, _ := json.Marshal(v)
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		var out []string
		for k := range m {
			out = append(out, k)
		}
		sort.Strings(out)
		return out
	}
	engine := bulk.Report{Errors: []bulk.RowError{{Row: 2, Column: "name", Message: "x"}}}
	doc := dto.ImportReport{Errors: []dto.ImportRowError{{Row: 2, Column: "name", Message: "x"}}}
	if !reflect.DeepEqual(keys(engine), keys(doc)) {
		t.Fatalf("report keys differ: engine %v, doc %v", keys(engine), keys(doc))
	}
	if !reflect.DeepEqual(keys(engine.Errors[0]), keys(doc.Errors[0])) {
		t.Fatalf("row error keys differ: engine %v, doc %v", keys(engine.Errors[0]), keys(doc.Errors[0]))
	}
}
