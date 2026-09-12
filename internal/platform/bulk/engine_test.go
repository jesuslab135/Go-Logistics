package bulk

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"fleet/internal/platform/apierr"
	"fleet/internal/platform/dbctx"
	"fleet/internal/platform/paginate"
)

func TestResolveRefPrefersNamesThenIDs(t *testing.T) {
	idx := newRefIndex([]Entry{{ID: 7, Name: "Unidad 101"}, {ID: 101, Name: "Tractor"}, {ID: 8, Name: "Duplicado"}, {ID: 9, Name: "duplicado"}})
	cases := []struct {
		cell, want, wantErr string
	}{
		{"unidad  101", "7", ""},
		{"101", "101", ""},
		{"7", "7", ""},
		{"999", "", `no activo with id 999 in this company`},
		{"Camión", "", `no activo named "Camión"`},
		{"DUPLICADO", "", `"DUPLICADO" matches 2 activo records; use the id instead`},
	}
	for _, tc := range cases {
		got, err := idx.resolve(tc.cell, "activo")
		if tc.wantErr != "" {
			if err == nil || err.Error() != tc.wantErr {
				t.Errorf("%q: err = %v, want %q", tc.cell, err, tc.wantErr)
			}
			continue
		}
		if err != nil || string(got) != tc.want {
			t.Errorf("%q: got %s, %v; want %s", tc.cell, got, err, tc.want)
		}
	}
}

func TestEntriesPagesThroughTheWholeList(t *testing.T) {
	all := make([]int64, 1203)
	for i := range all {
		all[i] = int64(i + 1)
	}
	list := func(_ context.Context, p paginate.Params) ([]int64, int64, error) {
		end := min(p.Offset+p.Limit, len(all))
		return all[p.Offset:end], int64(len(all)), nil
	}
	load := Entries(list, func(v int64) int64 { return v }, func(v int64) string { return "x" })
	got, err := load(context.Background())
	if err != nil || len(got) != len(all) {
		t.Fatalf("got %d entries, %v; want %d", len(got), err, len(all))
	}
}

func TestRowErrorsSpreadsFieldDetails(t *testing.T) {
	got := rowErrors(5, errors.New("boom"))
	if len(got) != 1 || got[0].Row != 5 || got[0].Message == "" {
		t.Fatalf("plain error: got %+v", got)
	}
}

// fakeTx is a minimal pgx.Tx: it tracks Begin/Commit/Rollback so a test can
// assert whether the outer transaction was committed or only ever rolled
// back, without a real database. Every other method is unused by the engine
// (real queries happen inside an Importer's own store, which these tests
// fake at a higher level) so they return zero values.
type fakeTx struct {
	beginErr   error
	commitErr  error
	committed  bool
	rolledBack bool
	children   []*fakeTx
}

func (t *fakeTx) Begin(context.Context) (pgx.Tx, error) {
	if t.beginErr != nil {
		return nil, t.beginErr
	}
	child := &fakeTx{}
	t.children = append(t.children, child)
	return child, nil
}
func (t *fakeTx) Commit(context.Context) error {
	t.committed = true
	return t.commitErr
}
func (t *fakeTx) Rollback(context.Context) error {
	t.rolledBack = true
	return nil
}
func (t *fakeTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (t *fakeTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (t *fakeTx) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (t *fakeTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (t *fakeTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (t *fakeTx) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (t *fakeTx) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (t *fakeTx) Conn() *pgx.Conn                                         { return nil }

// fakeImporter is a stand-in Importer: its Prepare is supplied by the test,
// so each scenario can decide what a "row" resolves to without a real store.
type fakeImporter struct {
	name    string
	cols    []Column
	refs    map[string]Lookup
	prepare func(values map[string]json.RawMessage) (func(ctx context.Context) error, map[string]string)
}

func (f *fakeImporter) Resource() string                          { return f.name }
func (f *fakeImporter) Columns(context.Context) ([]Column, error) { return f.cols, nil }
func (f *fakeImporter) Refs() map[string]Lookup                   { return f.refs }
func (f *fakeImporter) Prepare(values map[string]json.RawMessage) (func(ctx context.Context) error, map[string]string) {
	return f.prepare(values)
}

// cellString reads a plain string value assembled for a column, or "" when
// the row left it out (an empty optional cell).
func cellString(values map[string]json.RawMessage, key string) string {
	raw, ok := values[key]
	if !ok {
		return ""
	}
	var s string
	_ = json.Unmarshal(raw, &s)
	return s
}

func hasRowError(errs []RowError, want RowError) bool {
	for _, e := range errs {
		if e == want {
			return true
		}
	}
	return false
}

func TestRunReportsMissingRequiredColumnOnHeaderRow(t *testing.T) {
	imp := &fakeImporter{
		name: "activos",
		cols: []Column{{Key: "name", Kind: KindString, Required: true}, {Key: "code", Kind: KindString}},
		refs: map[string]Lookup{},
		prepare: func(map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			return func(context.Context) error { return nil }, nil
		},
	}
	rows := [][]string{{"code"}}

	report, err := Run(context.Background(), nil, imp, rows, false, Limits{})
	if err != nil {
		t.Fatalf("Run returned request-level error %v; want the problem in the report", err)
	}
	want := RowError{Row: 1, Column: "name", Message: "required column is missing"}
	if !hasRowError(report.Errors, want) {
		t.Fatalf("errors = %+v; want to contain %+v", report.Errors, want)
	}
	if report.Created != 0 {
		t.Fatalf("Created = %d; want 0", report.Created)
	}
}

func TestRunRejectsAnEmptyFile(t *testing.T) {
	imp := &fakeImporter{name: "activos", cols: []Column{}, refs: map[string]Lookup{}}

	_, err := Run(context.Background(), nil, imp, nil, false, Limits{})
	var apiErr *apierr.Error
	if !errors.As(err, &apiErr) || apiErr.Status != 400 || apiErr.Message != "the file is empty" {
		t.Fatalf("err = %v; want a 400 \"the file is empty\"", err)
	}
}

func TestRunEnforcesTheRowLimit(t *testing.T) {
	imp := &fakeImporter{
		name: "activos",
		cols: []Column{{Key: "name", Kind: KindString, Required: true}},
		refs: map[string]Lookup{},
		prepare: func(map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			return func(context.Context) error { return nil }, nil
		},
	}
	rows := [][]string{{"name"}, {"a"}, {"b"}, {"c"}}

	report, err := Run(context.Background(), nil, imp, rows, false, Limits{MaxRows: 2})
	var apiErr *apierr.Error
	if !errors.As(err, &apiErr) || apiErr.Status != 400 || !strings.Contains(apiErr.Message, "more than 2 rows") {
		t.Fatalf("err = %v; want a 400 mentioning the row limit", err)
	}
	if report.Created != 0 {
		t.Fatalf("Created = %d; want 0", report.Created)
	}
}

func TestRunIsAllOrNothingOnADatabaseRowError(t *testing.T) {
	root := &fakeTx{}
	ctx := dbctx.WithTx(context.Background(), root)
	var created []string
	imp := &fakeImporter{
		name: "activos",
		cols: []Column{{Key: "name", Kind: KindString, Required: true}},
		refs: map[string]Lookup{},
		prepare: func(values map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			name := cellString(values, "name")
			return func(context.Context) error {
				if name == "dup" {
					return apierr.Conflict("resource already exists")
				}
				created = append(created, name)
				return nil
			}, nil
		},
	}
	// Row 2 ("a") and row 4 ("b") are good; row 3 ("dup") fails in the
	// database. Nothing must be left committed, and the report must name
	// row 3 specifically (the header counts as row 1).
	rows := [][]string{{"name"}, {"a"}, {"dup"}, {"b"}}

	report, err := Run(ctx, nil, imp, rows, false, Limits{})
	if err != nil {
		t.Fatalf("Run returned request-level error %v; want the problem in the report", err)
	}
	if len(report.Errors) != 1 || report.Errors[0].Row != 3 || report.Errors[0].Message != "resource already exists" {
		t.Fatalf("errors = %+v; want exactly one error naming row 3", report.Errors)
	}
	if report.Created != 0 {
		t.Fatalf("Created = %d; want 0 because the import failed", report.Created)
	}
	// root carries no transaction of its own; dbctx.Begin nests Run's whole
	// import as one savepoint-backed transaction inside it (root.children[0]).
	// That is the "outer" transaction from Run's point of view, and it must
	// be rolled back, never committed, when any row fails.
	if len(root.children) != 1 {
		t.Fatalf("root has %d children; want exactly 1 (Run's own transaction)", len(root.children))
	}
	outer := root.children[0]
	if outer.committed {
		t.Fatal("outer transaction was committed; a failed import must leave the database untouched")
	}
	if !outer.rolledBack {
		t.Fatal("outer transaction was never rolled back")
	}
	if len(created) != 2 {
		t.Fatalf("store saw %d creates, want 2 (the good rows still run inside the doomed transaction)", len(created))
	}
}

func TestRunDryRunValidatesWithoutCommitting(t *testing.T) {
	root := &fakeTx{}
	ctx := dbctx.WithTx(context.Background(), root)
	imp := &fakeImporter{
		name: "activos",
		cols: []Column{{Key: "name", Kind: KindString, Required: true}},
		refs: map[string]Lookup{},
		prepare: func(map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			return func(context.Context) error { return nil }, nil
		},
	}
	rows := [][]string{{"name"}, {"a"}, {"b"}}

	report, err := Run(ctx, nil, imp, rows, true, Limits{})
	if err != nil {
		t.Fatalf("Run returned error %v", err)
	}
	if !report.DryRun || len(report.Errors) != 0 {
		t.Fatalf("report = %+v; want a clean dry run", report)
	}
	if report.Created != 2 {
		t.Fatalf("Created = %d; want 2 for a successful dry run", report.Created)
	}
	if len(root.children) != 1 {
		t.Fatalf("root has %d children; want exactly 1 (Run's own transaction)", len(root.children))
	}
	outer := root.children[0]
	if outer.committed {
		t.Fatal("a dry run must never commit the outer transaction")
	}
	if !outer.rolledBack {
		t.Fatal("a dry run must roll back what it ran")
	}
}

func TestRunFlagsUnknownAndAmbiguousReferenceNames(t *testing.T) {
	imp := &fakeImporter{
		name: "vueltas",
		cols: []Column{{Key: "asset", Kind: KindInt}},
		refs: map[string]Lookup{
			"asset": {Label: "activo", Load: func(context.Context) ([]Entry, error) {
				return []Entry{{ID: 1, Name: "Uno"}, {ID: 2, Name: "Dup"}, {ID: 3, Name: "Dup"}}, nil
			}},
		},
		prepare: func(map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			return func(context.Context) error { return nil }, nil
		},
	}
	rows := [][]string{{"asset"}, {"Nope"}, {"Dup"}, {"Uno"}}

	report, err := Run(context.Background(), nil, imp, rows, false, Limits{})
	if err != nil {
		t.Fatalf("Run returned request-level error %v; want the problem in the report", err)
	}
	wantUnknown := RowError{Row: 2, Column: "asset", Message: `no activo named "Nope"`}
	wantAmbiguous := RowError{Row: 3, Column: "asset", Message: `"Dup" matches 2 activo records; use the id instead`}
	if !hasRowError(report.Errors, wantUnknown) {
		t.Fatalf("errors = %+v; want %+v", report.Errors, wantUnknown)
	}
	if !hasRowError(report.Errors, wantAmbiguous) {
		t.Fatalf("errors = %+v; want %+v", report.Errors, wantAmbiguous)
	}
}

// TestRunSurfacesPerRowPrepareErrorsAsRowErrors covers the path an Importer
// uses to report a decode/validation/Assemble problem for one row (an
// Importer's Prepare, e.g. Flat.Prepare on an Assemble key collision, returns
// a non-nil map instead of a create func). Run must turn every entry into a
// RowError against the right spreadsheet row, sorted by column, and keep
// checking the remaining rows rather than aborting the request.
func TestRunSurfacesPerRowPrepareErrorsAsRowErrors(t *testing.T) {
	imp := &fakeImporter{
		name: "activos",
		cols: []Column{{Key: "name", Kind: KindString}},
		refs: map[string]Lookup{},
		prepare: func(values map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			if cellString(values, "name") == "bad" {
				// Mirrors what Flat.Prepare returns for an Assemble collision:
				// map[string]string{"": err.Error()}, plus a second offending
				// column to prove sorting.
				return nil, map[string]string{"zeta": "z problem", "": "collision"}
			}
			return func(context.Context) error { return nil }, nil
		},
	}
	rows := [][]string{{"name"}, {"good"}, {"bad"}}

	report, err := Run(context.Background(), nil, imp, rows, false, Limits{})
	if err != nil {
		t.Fatalf("Run returned request-level error %v; want the problem in the report", err)
	}
	want := []RowError{{Row: 3, Message: "collision"}, {Row: 3, Column: "zeta", Message: "z problem"}}
	if len(report.Errors) != len(want) {
		t.Fatalf("errors = %+v; want %+v", report.Errors, want)
	}
	for i := range want {
		if report.Errors[i] != want[i] {
			t.Fatalf("errors[%d] = %+v; want %+v (sorted by column)", i, report.Errors[i], want[i])
		}
	}
}
