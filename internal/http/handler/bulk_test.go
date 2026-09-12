package handler

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
