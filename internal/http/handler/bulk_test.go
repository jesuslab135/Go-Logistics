package handler

import (
	"bytes"
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
