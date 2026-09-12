package bulk

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
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
	// No per-row savepoints. One savepoint per row spent a subtransaction XID
	// per row, and a Postgres backend caches only 64 of them before every
	// other backend's visibility checks start hitting pg_subtrans - certain to
	// happen at the 5000-row cap. They also bought nothing: the import is
	// all-or-nothing, so a row that fails in the database ends the run rather
	// than being stepped over.
	if len(outer.children) != 0 {
		t.Fatalf("outer opened %d savepoints; want 0 (the per-row savepoint loop is gone)", len(outer.children))
	}
	// The run stops at the first database-level failure, so row 4 ("b") is
	// never attempted. "a" still ran, inside the transaction about to be
	// rolled back.
	if len(created) != 1 || created[0] != "a" {
		t.Fatalf("store saw creates %v; want exactly [a] - the run must stop at the first failing row", created)
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

// TestRunTreatsInfrastructureFailuresAsRequestLevel covers a create call that
// fails with a plain error apierr.Map cannot recognise (a dropped connection,
// a context cancellation, a statement timeout - not a unique violation, not a
// validation problem). That is not this row's data being wrong; it is the
// database failing under the import. Run must surface it as a request-level
// error, once, rather than folding it into a "internal server error" row
// entry and letting the loop hammer every remaining row with the same
// useless message under a 200 response.
func TestRunTreatsInfrastructureFailuresAsRequestLevel(t *testing.T) {
	root := &fakeTx{}
	ctx := dbctx.WithTx(context.Background(), root)
	var attempts int
	imp := &fakeImporter{
		name: "activos",
		cols: []Column{{Key: "name", Kind: KindString, Required: true}},
		refs: map[string]Lookup{},
		prepare: func(values map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			name := cellString(values, "name")
			return func(context.Context) error {
				attempts++
				if name == "boom" {
					return errors.New("connection reset by peer")
				}
				return nil
			}, nil
		},
	}
	rows := [][]string{{"name"}, {"a"}, {"boom"}, {"c"}}

	report, err := Run(ctx, nil, imp, rows, false, Limits{})
	var apiErr *apierr.Error
	if !errors.As(err, &apiErr) || apiErr.Status < 500 {
		t.Fatalf("err = %v; want a 5xx apierr.Error", err)
	}
	if len(report.Errors) != 0 {
		t.Fatalf("errors = %+v; want none - this is not a row problem", report.Errors)
	}
	if report.Created != 0 {
		t.Fatalf("Created = %d; want 0", report.Created)
	}
	if attempts != 2 {
		t.Fatalf("create was attempted %d times; want 2 (row \"c\" must not run after the infrastructure failure)", attempts)
	}
	if len(root.children) != 1 {
		t.Fatalf("root has %d children; want exactly 1 (Run's own transaction)", len(root.children))
	}
	if root.children[0].committed {
		t.Fatal("outer transaction was committed; an infrastructure failure must leave the database untouched")
	}
	if !root.children[0].rolledBack {
		t.Fatal("outer transaction was never rolled back")
	}
}

// TestRunStillRecordsGenuineDataErrorsAsRowErrors guards against the F1 fix
// over-firing: a unique violation (409) and other ordinary 4xx problems are
// this row's data being wrong, not an outage, and must stay row errors with
// the loop continuing - exactly the behavior already proven by
// TestRunIsAllOrNothingOnADatabaseRowError, restated here for a plain
// validation-shaped 4xx rather than a Postgres conflict.
func TestRunStillRecordsGenuineDataErrorsAsRowErrors(t *testing.T) {
	root := &fakeTx{}
	ctx := dbctx.WithTx(context.Background(), root)
	imp := &fakeImporter{
		name: "activos",
		cols: []Column{{Key: "name", Kind: KindString, Required: true}},
		refs: map[string]Lookup{},
		prepare: func(values map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			name := cellString(values, "name")
			return func(context.Context) error {
				if name == "bad" {
					return apierr.BadRequest("bad value")
				}
				return nil
			}, nil
		},
	}
	rows := [][]string{{"name"}, {"a"}, {"bad"}, {"c"}}

	report, err := Run(ctx, nil, imp, rows, false, Limits{})
	if err != nil {
		t.Fatalf("Run returned request-level error %v; a 4xx row problem must stay in the report", err)
	}
	if len(report.Errors) != 1 || report.Errors[0].Row != 3 || report.Errors[0].Message != "bad value" {
		t.Fatalf("errors = %+v; want exactly one error naming row 3", report.Errors)
	}
}

func TestRunFlagsDuplicateHeaderColumns(t *testing.T) {
	imp := &fakeImporter{
		name: "activos",
		cols: []Column{{Key: "name", Kind: KindString}, {Key: "code", Kind: KindString}},
		refs: map[string]Lookup{},
		prepare: func(map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			return func(context.Context) error { return nil }, nil
		},
	}
	rows := [][]string{{"name", "code", "name"}, {"a", "1", "b"}}

	report, err := Run(context.Background(), nil, imp, rows, false, Limits{})
	if err != nil {
		t.Fatalf("Run returned request-level error %v; want the problem in the report", err)
	}
	want := RowError{Row: 1, Column: "name", Message: "column is duplicated"}
	if !hasRowError(report.Errors, want) {
		t.Fatalf("errors = %+v; want to contain %+v", report.Errors, want)
	}
	if report.Created != 0 {
		t.Fatalf("Created = %d; want 0", report.Created)
	}
}

func TestRunBlankRowInTheMiddleDoesNotSkewCountOrRowNumbers(t *testing.T) {
	root := &fakeTx{}
	ctx := dbctx.WithTx(context.Background(), root)
	imp := &fakeImporter{
		name: "activos",
		cols: []Column{{Key: "name", Kind: KindString}},
		refs: map[string]Lookup{},
		prepare: func(values map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			name := cellString(values, "name")
			return func(context.Context) error {
				if name == "bad" {
					return apierr.BadRequest("bad value")
				}
				return nil
			}, nil
		},
	}
	// Spreadsheet rows: 1 header, 2 "a" (good), 3 blank, 4 "bad" (fails).
	// Blank must not be counted and must not shift row 4's own number.
	rows := [][]string{{"name"}, {"a"}, {""}, {"bad"}}

	report, err := Run(ctx, nil, imp, rows, false, Limits{})
	if err != nil {
		t.Fatalf("Run returned request-level error %v; want the problem in the report", err)
	}
	if report.Rows != 2 {
		t.Fatalf("Rows = %d; want 2 (the blank row must not be counted)", report.Rows)
	}
	want := RowError{Row: 4, Message: "bad value"}
	if !hasRowError(report.Errors, want) {
		t.Fatalf("errors = %+v; want %+v (true spreadsheet row, unaffected by the blank row)", report.Errors, want)
	}
}

// Unknown headers stay ignored - that is what lets a file exported from a list
// be re-uploaded unchanged, id and timestamps and all - but a MISSPELLED
// optional column is indistinguishable from one of those and used to import as
// empty under a clean 201, leaving the user sure the data had landed. Report
// them rather than reject them, so both cases keep working.
func TestRunReportsIgnoredColumns(t *testing.T) {
	imp := &fakeImporter{
		name: "activos",
		cols: []Column{{Key: "name", Kind: KindString}, {Key: "code", Kind: KindString}},
		refs: map[string]Lookup{},
		prepare: func(map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			return func(context.Context) error { return nil }, nil
		},
	}
	// "id" is a column an exported file carries; "nmae" is a typo for "name";
	// the trailing blank header is not a column at all and must not be listed.
	rows := [][]string{{"name", "id", "nmae", ""}, {"a", "7", "b", ""}}

	// This file has no row errors, so Run reaches the transaction: give it one
	// to nest in rather than a nil pool.
	ctx := dbctx.WithTx(context.Background(), &fakeTx{})
	report, err := Run(ctx, nil, imp, rows, false, Limits{})
	if err != nil {
		t.Fatalf("Run returned request-level error %v", err)
	}
	if len(report.Errors) != 0 {
		t.Fatalf("errors = %+v; an unknown column must stay non-fatal", report.Errors)
	}
	want := []string{"id", "nmae"}
	if !reflect.DeepEqual(report.IgnoredColumns, want) {
		t.Fatalf("IgnoredColumns = %v; want %v", report.IgnoredColumns, want)
	}
}

// Errors stops at maxErrors, so the report has to say how many problems there
// really were: otherwise a user fixes the 500 listed, re-uploads, and is met
// with 500 more with nothing having warned them there were ever others.
func TestRunCountsEveryErrorBeyondTheListedCap(t *testing.T) {
	imp := &fakeImporter{
		name: "activos",
		cols: []Column{{Key: "age", Kind: KindInt}},
		refs: map[string]Lookup{},
		prepare: func(map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			return func(context.Context) error { return nil }, nil
		},
	}
	const extra = 120
	rows := [][]string{{"age"}}
	for i := 0; i < maxErrors+extra; i++ {
		rows = append(rows, []string{"not-a-number"})
	}

	report, err := Run(context.Background(), nil, imp, rows, false, Limits{})
	if err != nil {
		t.Fatalf("Run returned request-level error %v", err)
	}
	if len(report.Errors) != maxErrors {
		t.Fatalf("listed errors = %d; want the %d cap", len(report.Errors), maxErrors)
	}
	if report.ErrorCount != maxErrors+extra {
		t.Fatalf("ErrorCount = %d; want %d (every problem counted, not just the listed ones)", report.ErrorCount, maxErrors+extra)
	}
	if !report.Truncated {
		t.Fatal("Truncated = false; the listed errors are only a prefix of what was found")
	}
}

// A clean report must not claim truncation, or the flag means nothing.
func TestRunLeavesTruncatedUnsetForASmallReport(t *testing.T) {
	imp := &fakeImporter{
		name: "activos",
		cols: []Column{{Key: "age", Kind: KindInt}},
		refs: map[string]Lookup{},
		prepare: func(map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			return func(context.Context) error { return nil }, nil
		},
	}
	rows := [][]string{{"age"}, {"nope"}}

	report, err := Run(context.Background(), nil, imp, rows, false, Limits{})
	if err != nil {
		t.Fatalf("Run returned request-level error %v", err)
	}
	if report.ErrorCount != 1 || report.Truncated {
		t.Fatalf("ErrorCount = %d, Truncated = %v; want 1 and false", report.ErrorCount, report.Truncated)
	}
}

func TestRunReportsColumnScopedErrorForABadCellConversion(t *testing.T) {
	imp := &fakeImporter{
		name: "activos",
		cols: []Column{{Key: "name", Kind: KindString}, {Key: "age", Kind: KindInt}},
		refs: map[string]Lookup{},
		prepare: func(map[string]json.RawMessage) (func(context.Context) error, map[string]string) {
			return func(context.Context) error { return nil }, nil
		},
	}
	rows := [][]string{{"name", "age"}, {"a", "30"}, {"", ""}, {"bad", "not-a-number"}}

	report, err := Run(context.Background(), nil, imp, rows, false, Limits{})
	if err != nil {
		t.Fatalf("Run returned request-level error %v; want the problem in the report", err)
	}
	want := RowError{Row: 4, Column: "age", Message: "must be a whole number"}
	if !hasRowError(report.Errors, want) {
		t.Fatalf("errors = %+v; want %+v", report.Errors, want)
	}
	if report.Rows != 2 {
		t.Fatalf("Rows = %d; want 2 (the blank row must not be counted)", report.Rows)
	}
}
