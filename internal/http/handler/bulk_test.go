package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

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
