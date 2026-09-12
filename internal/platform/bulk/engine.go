package bulk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/domain/customfield"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/crud"
	"fleet/internal/platform/dbctx"
	"fleet/internal/platform/paginate"
)

const maxErrors = 500

// RowError is one problem, located by spreadsheet row (the header is row 1).
type RowError struct {
	Row     int    `json:"row"`
	Column  string `json:"column,omitempty"`
	Message string `json:"message"`
}

// Report is the outcome of an import. Created is 0 unless it was committed or
// is a successful dry run.
type Report struct {
	Resource string     `json:"resource"`
	DryRun   bool       `json:"dry_run"`
	Rows     int        `json:"rows"`
	Created  int        `json:"created"`
	Errors   []RowError `json:"errors"`
}

// Entry is one record a reference column can point at.
type Entry struct {
	ID   int64
	Name string
}

// Lookup is what a reference column resolves through. Label names the
// referenced thing in messages and the template ("activo").
type Lookup struct {
	Label string
	Load  func(ctx context.Context) ([]Entry, error)
}

// Entries pages through a store's List and keeps each record's id and name.
// Stores scope List to the caller's company, so lookups can only ever match
// the caller's own records.
func Entries[T any](list func(context.Context, paginate.Params) ([]T, int64, error), id func(T) int64, name func(T) string) func(context.Context) ([]Entry, error) {
	return func(ctx context.Context) ([]Entry, error) {
		var out []Entry
		for offset := 0; ; {
			items, total, err := list(ctx, paginate.Params{Limit: 500, Offset: offset})
			if err != nil {
				return nil, err
			}
			for _, it := range items {
				out = append(out, Entry{ID: id(it), Name: name(it)})
			}
			offset += len(items)
			if len(items) == 0 || int64(offset) >= total {
				return out, nil
			}
		}
	}
}

// Importer is one section's import.
type Importer interface {
	Resource() string
	Columns(ctx context.Context) ([]Column, error)
	Refs() map[string]Lookup
	// Prepare decodes and validates one row. It returns the create to run, or
	// the problems found, keyed by column.
	Prepare(values map[string]json.RawMessage) (func(ctx context.Context) error, map[string]string)
}

// Flat imports a top-level section through its crud.Store.
type Flat[T, C, U any] struct {
	Name  string
	Store crud.Store[T, C, U]
	// CustomFields returns the company's custom field definitions for this
	// section; nil when the section has none.
	CustomFields func(ctx context.Context) ([]customfield.Definition, error)
	References   map[string]Lookup
	Skip         []string
	Kinds        map[string]Kind // overrides SchemaOf's kind for a key
}

func (f Flat[T, C, U]) Resource() string        { return f.Name }
func (f Flat[T, C, U]) Refs() map[string]Lookup { return f.References }

func (f Flat[T, C, U]) Columns(ctx context.Context) ([]Column, error) {
	cols := decorate(SchemaOf(reflect.TypeFor[C](), f.Skip...), f.References, f.Kinds)
	if f.CustomFields == nil {
		return cols, nil
	}
	defs, err := f.CustomFields(ctx)
	if err != nil {
		return nil, err
	}
	for _, d := range defs {
		col := Column{Key: "cf." + d.Key, Kind: KindString, Required: d.Required, Label: d.Label, CustomType: d.Type}
		if d.Type == customfield.TypeSelect {
			col.OneOf = d.Options
		}
		cols = append(cols, col)
	}
	return cols, nil
}

func (f Flat[T, C, U]) Prepare(values map[string]json.RawMessage) (func(ctx context.Context) error, map[string]string) {
	doc, err := Assemble(values)
	if err != nil {
		return nil, map[string]string{"": err.Error()}
	}
	in, errs := Decode[C](doc)
	if errs != nil {
		return nil, errs
	}
	return func(ctx context.Context) error {
		_, err := f.Store.Create(ctx, in)
		return err
	}, nil
}

// Nested imports a child section through its crud.NestedStore. ParentColumn
// holds the parent, resolved through References like any reference.
type Nested[T, C, U any] struct {
	Name         string
	Store        crud.NestedStore[T, C, U]
	ParentColumn string
	References   map[string]Lookup
	Skip         []string
	Kinds        map[string]Kind
}

func (n Nested[T, C, U]) Resource() string        { return n.Name }
func (n Nested[T, C, U]) Refs() map[string]Lookup { return n.References }

func (n Nested[T, C, U]) Columns(context.Context) ([]Column, error) {
	parent := Column{Key: n.ParentColumn, Kind: KindInt, Required: true}
	cols := append([]Column{parent}, SchemaOf(reflect.TypeFor[C](), n.Skip...)...)
	return decorate(cols, n.References, n.Kinds), nil
}

func (n Nested[T, C, U]) Prepare(values map[string]json.RawMessage) (func(ctx context.Context) error, map[string]string) {
	parentID, err := strconv.ParseInt(string(values[n.ParentColumn]), 10, 64)
	if err != nil {
		return nil, map[string]string{n.ParentColumn: "this field is required"}
	}
	delete(values, n.ParentColumn)
	doc, err := Assemble(values)
	if err != nil {
		return nil, map[string]string{"": err.Error()}
	}
	in, errs := Decode[C](doc)
	if errs != nil {
		return nil, errs
	}
	return func(ctx context.Context) error {
		_, err := n.Store.Create(ctx, parentID, in)
		return err
	}, nil
}

func decorate(cols []Column, refs map[string]Lookup, kinds map[string]Kind) []Column {
	for i := range cols {
		if k, ok := kinds[cols[i].Key]; ok {
			cols[i].Kind = k
		}
		if lk, ok := refs[cols[i].Key]; ok {
			cols[i].Ref = cols[i].Key
			cols[i].Label = lk.Label
		}
	}
	return cols
}

// Limits bounds one import.
type Limits struct{ MaxRows int }

type prepared struct {
	row    int
	create func(ctx context.Context) error
}

// Run imports rows (row 0 is the header) all-or-nothing. It returns an error
// only for problems with the request itself; problems with rows are in the
// report.
func Run(ctx context.Context, pool *pgxpool.Pool, imp Importer, rows [][]string, dryRun bool, limits Limits) (Report, error) {
	report := Report{Resource: imp.Resource(), DryRun: dryRun, Errors: []RowError{}}
	if len(rows) == 0 {
		return report, apierr.BadRequest("the file is empty")
	}
	cols, err := imp.Columns(ctx)
	if err != nil {
		return report, err
	}

	byKey := make(map[string]Column, len(cols))
	for _, c := range cols {
		byKey[c.Key] = c
	}
	position := map[int]Column{}
	present := map[string]bool{}
	duplicate := map[string]bool{}
	for i, h := range rows[0] {
		c, ok := byKey[strings.TrimSpace(h)]
		if !ok {
			continue
		}
		if present[c.Key] {
			if !duplicate[c.Key] {
				report.add(RowError{Row: 1, Column: c.Key, Message: "column is duplicated"})
				duplicate[c.Key] = true
			}
			continue
		}
		position[i] = c
		present[c.Key] = true
	}
	for _, c := range cols {
		if c.Required && !present[c.Key] {
			report.add(RowError{Row: 1, Column: c.Key, Message: "required column is missing"})
		}
	}

	indexes := map[string]*refIndex{}
	for key, lk := range imp.Refs() {
		if !present[key] {
			continue
		}
		entries, err := lk.Load(ctx)
		if err != nil {
			return report, err
		}
		indexes[key] = newRefIndex(entries)
	}

	var ready []prepared
	for r := 1; r < len(rows); r++ {
		if blank(rows[r]) {
			continue
		}
		report.Rows++
		if limits.MaxRows > 0 && report.Rows > limits.MaxRows {
			return report, apierr.BadRequest(fmt.Sprintf("the file has more than %d rows; split it", limits.MaxRows))
		}
		rowNo := r + 1
		values := map[string]json.RawMessage{}
		ok := true
		for i, cell := range rows[r] {
			col, known := position[i]
			if !known {
				continue
			}
			var v json.RawMessage
			var err error
			if idx, isRef := indexes[col.Key]; isRef && strings.TrimSpace(cell) != "" {
				v, err = idx.resolve(cell, imp.Refs()[col.Key].Label)
			} else {
				v, err = CellValue(col, cell)
			}
			if err != nil {
				report.add(RowError{Row: rowNo, Column: col.Key, Message: err.Error()})
				ok = false
				continue
			}
			if v != nil {
				values[col.Key] = v
			}
		}
		if !ok {
			continue
		}
		create, errs := imp.Prepare(values)
		if errs != nil {
			for _, e := range sortedErrors(rowNo, errs) {
				report.add(e)
			}
			continue
		}
		ready = append(ready, prepared{row: rowNo, create: create})
	}
	if len(report.Errors) > 0 {
		return report, nil
	}

	tx, err := dbctx.Begin(ctx, pool)
	if err != nil {
		return report, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // a no-op after Commit
	txCtx := dbctx.WithTx(ctx, tx)

	for _, p := range ready {
		sp, err := tx.Begin(txCtx)
		if err != nil {
			return report, err
		}
		if err := p.create(txCtx); err != nil {
			_ = sp.Rollback(txCtx)
			mapped := apierr.Map(err)
			// A 5xx here is not a problem with this row's data (those are
			// already 4xx: a unique violation, a bad foreign key, a failed
			// validation) - it is the database or connection failing under
			// us. Recording it as one more row error would flood the report
			// with up to 500 identical "internal server error" entries,
			// answer 200, and hide the outage from anything watching Run's
			// own return value. Treat it as a request-level failure instead;
			// the deferred rollback still undoes everything already done.
			if mapped.Status >= 500 {
				report.Created = 0
				return report, mapped
			}
			for _, e := range mappedRowErrors(p.row, mapped) {
				report.add(e)
			}
			continue
		}
		if err := sp.Commit(txCtx); err != nil {
			return report, err
		}
		report.Created++
	}

	if len(report.Errors) > 0 {
		report.Created = 0
		return report, nil
	}
	if dryRun {
		return report, nil
	}
	if err := tx.Commit(ctx); err != nil {
		return report, err
	}
	return report, nil
}

func (r *Report) add(e RowError) {
	if len(r.Errors) < maxErrors {
		r.Errors = append(r.Errors, e)
	}
}

func blank(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func sortedErrors(row int, errs map[string]string) []RowError {
	out := make([]RowError, 0, len(errs))
	for col, msg := range errs {
		out = append(out, RowError{Row: row, Column: col, Message: msg})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Column < out[j].Column })
	return out
}

// rowErrors turns a store error into row errors, keeping per-field details
// when the error carries them.
func rowErrors(row int, err error) []RowError {
	return mappedRowErrors(row, apierr.Map(err))
}

// mappedRowErrors is rowErrors' second half, taking an already-mapped error
// so a caller that must inspect the mapping first (to tell a data problem
// from an infrastructure one) does not map it twice.
func mappedRowErrors(row int, e *apierr.Error) []RowError {
	if details, ok := e.Details.(map[string]string); ok && len(details) > 0 {
		return sortedErrors(row, details)
	}
	return []RowError{{Row: row, Message: e.Message}}
}

type refIndex struct {
	byName map[string][]int64
	ids    map[int64]bool
}

func newRefIndex(entries []Entry) *refIndex {
	idx := &refIndex{byName: map[string][]int64{}, ids: map[int64]bool{}}
	for _, e := range entries {
		n := normalize(e.Name)
		idx.byName[n] = append(idx.byName[n], e.ID)
		idx.ids[e.ID] = true
	}
	return idx
}

// resolve matches a name first and an id second, so a record named "101" is
// found by its name even when some other record has id 101.
func (idx *refIndex) resolve(cell, label string) (json.RawMessage, error) {
	switch ids := idx.byName[normalize(cell)]; len(ids) {
	case 1:
		return json.RawMessage(strconv.FormatInt(ids[0], 10)), nil
	case 0:
	default:
		return nil, fmt.Errorf("%q matches %d %s records; use the id instead", cell, len(ids), label)
	}
	if id, err := strconv.ParseInt(strings.TrimSpace(cell), 10, 64); err == nil {
		if idx.ids[id] {
			return json.RawMessage(strconv.FormatInt(id, 10)), nil
		}
		return nil, fmt.Errorf("no %s with id %d in this company", label, id)
	}
	return nil, errors.New("no " + label + " named " + strconv.Quote(cell))
}

func normalize(s string) string { return strings.Join(strings.Fields(strings.ToLower(s)), " ") }
