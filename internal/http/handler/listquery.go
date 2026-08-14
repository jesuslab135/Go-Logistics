package handler

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/platform/apierr"
	"fleet/internal/platform/filter"
	"fleet/internal/platform/paginate"
)

// sqlc generates one fixed SQL string per query, so a resource whose filters
// vary per request cannot be served from it. The types below assemble such a
// statement instead, under three rules that hold by construction rather than by
// convention:
//
//  1. Nothing from the request is ever concatenated into SQL. Table and alias
//     names are validated identifiers, join and select expressions are rejected
//     if they could terminate a statement, order columns come from an
//     allow-list, and every value is a bound parameter.
//  2. The SELECT names T's columns explicitly, in T's field order, so scanning
//     positionally is correct by definition instead of relying on "*" happening
//     to match the struct.
//  3. The count query reuses the same FROM and WHERE, so the reported total
//     always describes the filtered set.

var identifierPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// mustIdentifier guards the only two strings that reach SQL by concatenation
// without being expressions. A bad value here is a wiring mistake, not user
// input, so it fails loudly at the call site.
func mustIdentifier(kind, s string) string {
	if !identifierPattern.MatchString(s) {
		panic(fmt.Sprintf("listquery: %s %q is not a plain identifier", kind, s))
	}
	return s
}

// mustClause guards code-owned SQL fragments (joins, select expressions). They
// are never built from request data; this only stops a fragment from carrying a
// statement terminator into the query.
func mustClause(kind, s string) string {
	if strings.ContainsAny(s, ";") {
		panic(fmt.Sprintf("listquery: %s %q must not contain a statement terminator", kind, s))
	}
	return s
}

var columnCache sync.Map // reflect.Type -> []string

// tableColumns returns the column names of a sqlc row struct in field order.
// sqlc derives Go field names from the column names, so reversing that mapping
// reproduces the table's columns without hardcoding them anywhere.
func tableColumns(t reflect.Type) []string {
	if cached, ok := columnCache.Load(t); ok {
		return cached.([]string)
	}
	cols := make([]string, 0, t.NumField())
	for i := range t.NumField() {
		cols = append(cols, snakeCase(t.Field(i).Name))
	}
	columnCache.Store(t, cols)
	return cols
}

// snakeCase reverses sqlc's column -> field naming: CompanyID becomes
// company_id, TreadDepth32nds becomes tread_depth_32nds.
func snakeCase(name string) string {
	var b strings.Builder
	b.Grow(len(name) + 4)
	runes := []rune(name)
	for i, r := range runes {
		if i > 0 && (unicode.IsUpper(r) || unicode.IsDigit(r)) {
			prev := runes[i-1]
			if unicode.IsLower(prev) || (unicode.IsDigit(prev) && unicode.IsUpper(r)) {
				b.WriteByte('_')
			}
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// listSpec is a filtered list query over row type T.
//
// T is either a sqlc row struct, or a wrapper whose first field embeds one and
// whose remaining fields are supplied by extras — one expression per field, in
// order. Any mismatch between the two panics at build time rather than
// silently scanning a column into the wrong field.
type listSpec[T any] struct {
	table  string   // base table name
	alias  string   // its alias, used to qualify T's columns
	joins  []string // code-owned JOIN clauses
	extras []string // select expressions for T's trailing, non-embedded fields
	where  *filter.Where
	order  string // rendered by orderClause, so already allow-listed

	// countDistinct switches the count to count(DISTINCT <alias>.id), for the
	// case where a join multiplies rows. Plain count(*) is correct for the
	// 1:1 and N:1 joins used today.
	countDistinct bool
}

func newListSpec[T any](table, alias string) listSpec[T] {
	return listSpec[T]{table: mustIdentifier("table", table), alias: mustIdentifier("alias", alias)}
}

func (s listSpec[T]) join(clauses ...string) listSpec[T] {
	for _, c := range clauses {
		s.joins = append(s.joins, mustClause("join", c))
	}
	return s
}

func (s listSpec[T]) selecting(extras ...string) listSpec[T] {
	for _, e := range extras {
		s.extras = append(s.extras, mustClause("select expression", e))
	}
	return s
}

func (s listSpec[T]) filter(w *filter.Where) listSpec[T] {
	s.where = w
	return s
}

func (s listSpec[T]) orderBy(order string) listSpec[T] {
	s.order = order
	return s
}

func (s listSpec[T]) distinct() listSpec[T] {
	s.countDistinct = true
	return s
}

// selectList renders the projection: T's embedded table columns qualified by
// the alias, followed by the caller's extra expressions.
func (s listSpec[T]) selectList() string {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("listquery: row type %s is not a struct", t))
	}

	var exprs []string
	if t.NumField() > 0 && t.Field(0).Anonymous && t.Field(0).Type.Kind() == reflect.Struct {
		// Wrapper: embedded row first, then one expression per extra field.
		embedded := t.Field(0).Type
		exprs = make([]string, 0, embedded.NumField()+t.NumField()-1)
		for _, col := range tableColumns(embedded) {
			exprs = append(exprs, s.alias+"."+col)
		}
		if want := t.NumField() - 1; want != len(s.extras) {
			panic(fmt.Sprintf("listquery: %s declares %d field(s) beyond the embedded row but %d select expression(s) were given",
				t, want, len(s.extras)))
		}
		exprs = append(exprs, s.extras...)
	} else {
		if len(s.extras) > 0 {
			panic(fmt.Sprintf("listquery: %s embeds no row struct, so it cannot take extra select expressions", t))
		}
		exprs = make([]string, 0, t.NumField())
		for _, col := range tableColumns(t) {
			exprs = append(exprs, s.alias+"."+col)
		}
	}
	return "SELECT " + strings.Join(exprs, ", ")
}

func (s listSpec[T]) fromClause() string {
	from := "FROM " + s.table + " " + s.alias
	if len(s.joins) > 0 {
		from += " " + strings.Join(s.joins, " ")
	}
	return from
}

// build renders the page query and the matching count query.
func (s listSpec[T]) build(p paginate.Params) (string, []any, string, []any) {
	where := s.where
	if where == nil {
		where = filter.NewWhere(1)
	}

	cond := ""
	if w := where.SQL(); w != "" {
		cond = " WHERE " + w
	}
	order := ""
	if s.order != "" {
		order = " ORDER BY " + s.order
	}

	n := where.Next()
	listSQL := fmt.Sprintf("%s %s%s%s LIMIT $%d OFFSET $%d", s.selectList(), s.fromClause(), cond, order, n, n+1)
	listArgs := append(append([]any{}, where.Args()...), p.Limit, p.Offset)

	countExpr := "count(*)"
	if s.countDistinct {
		countExpr = "count(DISTINCT " + s.alias + ".id)"
	}
	countSQL := "SELECT " + countExpr + " " + s.fromClause() + cond
	return listSQL, listArgs, countSQL, where.Args()
}

// runList executes a spec and returns the page plus the filtered total.
func runList[T any](ctx context.Context, pool *pgxpool.Pool, s listSpec[T], p paginate.Params) ([]T, int64, error) {
	listSQL, listArgs, countSQL, countArgs := s.build(p)

	rows, err := pool.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	items, err := pgx.CollectRows(rows, pgx.RowToStructByPos[T])
	if err != nil {
		return nil, 0, err
	}

	var total int64
	if err := pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// The query-parameter helpers below all treat an absent or empty value as "no
// filter" and a malformed one as a 400, so a typo never silently returns the
// unfiltered collection.

func queryInt64(c *gin.Context, name string) (*int64, error) {
	raw := c.Query(name)
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, apierr.BadRequest(fmt.Sprintf("invalid %s", name))
	}
	return &v, nil
}

func queryStr(c *gin.Context, name string) *string {
	raw := c.Query(name)
	if raw == "" {
		return nil
	}
	return &raw
}

func queryBool(c *gin.Context, name string) (*bool, error) {
	raw := c.Query(name)
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, apierr.BadRequest(fmt.Sprintf("invalid %s: use true or false", name))
	}
	return &v, nil
}

// queryTime accepts a full RFC3339 timestamp or a bare date. A bare date is
// midnight UTC, so a "_to" bound expressed that way excludes the named day.
func queryTime(c *gin.Context, name string) (*time.Time, error) {
	raw := c.Query(name)
	if raw == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t, nil
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return &t, nil
	}
	return nil, apierr.BadRequest(fmt.Sprintf("invalid %s: use RFC3339 or YYYY-MM-DD", name))
}

// queryEnum rejects values outside the documented vocabulary rather than
// returning an empty page for a misspelled state.
func queryEnum(c *gin.Context, name string, allowed ...string) (*string, error) {
	raw := c.Query(name)
	if raw == "" {
		return nil, nil
	}
	if slices.Contains(allowed, raw) {
		return &raw, nil
	}
	return nil, apierr.BadRequest(fmt.Sprintf("invalid %s: must be one of %s", name, strings.Join(allowed, ", ")))
}

// orderClause renders ?order= (comma-separated, "-" prefix for descending)
// against an allow-list, then appends idCol so the ordering is total. Without
// that tie-breaker, rows equal on the sort key can repeat or disappear between
// pages.
func orderClause(c *gin.Context, allowed map[string]string, fallback []filter.Sort, idCol string) string {
	sorts := filter.ParseSort(c.Query("order"), allowed, fallback)
	desc := false
	if n := len(sorts); n > 0 {
		desc = sorts[n-1].Desc
	}
	return filter.OrderBy(append(slices.Clone(sorts), filter.Sort{Column: idCol, Desc: desc}))
}
