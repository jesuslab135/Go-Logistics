package filter

import (
	"fmt"
	"strings"
)

// Op is a whitelisted SQL comparison operator. Only values defined here can
// ever reach a query, so operators are never taken from user input directly.
type Op string

const (
	Eq   Op = "="
	Neq  Op = "<>"
	Lt   Op = "<"
	Lte  Op = "<="
	Gt   Op = ">"
	Gte  Op = ">="
	Like Op = "ILIKE"
	In   Op = "IN"
)

// Field binds a request parameter name to a real column + operator. Callers
// build an allow-list of these; column names therefore never come from input.
type Field struct {
	Column string
	Op     Op
}

// Where accumulates parameterized conditions. It tracks the next placeholder
// index so it can be appended after fixed conditions (e.g. a tenant scope that
// already consumed $1).
type Where struct {
	next    int
	clauses []string
	args    []any
}

// NewWhere starts a builder whose first placeholder is $start. Pass the number
// of parameters the caller has already bound plus one.
func NewWhere(start int) *Where {
	if start < 1 {
		start = 1
	}
	return &Where{next: start}
}

func (w *Where) Add(column string, op Op, value any) *Where {
	if op == In {
		w.clauses = append(w.clauses, fmt.Sprintf("%s = ANY($%d)", column, w.next))
	} else {
		w.clauses = append(w.clauses, fmt.Sprintf("%s %s $%d", column, op, w.next))
	}
	w.args = append(w.args, value)
	w.next++
	return w
}

// Raw appends a hand-written condition for shapes Add cannot express, such as
// an OR group across several columns. Each "?" becomes the next positional
// placeholder, so the clause stays parameterized; args must match the number of
// "?" occurrences. The clause itself is code-owned — never request input.
func (w *Where) Raw(clause string, args ...any) *Where {
	var b strings.Builder
	for _, r := range clause {
		if r == '?' {
			fmt.Fprintf(&b, "$%d", w.next)
			w.next++
			continue
		}
		b.WriteRune(r)
	}
	w.clauses = append(w.clauses, b.String())
	w.args = append(w.args, args...)
	return w
}

// FromQuery adds one condition per (param -> Field) match found in values.
// Params absent from allowed are ignored, so unknown or malicious keys are
// silently dropped rather than reaching SQL.
func (w *Where) FromQuery(values map[string]string, allowed map[string]Field) *Where {
	for key, f := range allowed {
		if raw, ok := values[key]; ok && raw != "" {
			if f.Op == Like {
				raw = "%" + raw + "%"
			}
			w.Add(f.Column, f.Op, raw)
		}
	}
	return w
}

// SQL returns the joined conditions without the WHERE keyword, or "" if empty.
func (w *Where) SQL() string {
	return strings.Join(w.clauses, " AND ")
}

func (w *Where) Args() []any { return w.args }

// Next is the placeholder index a caller should continue from (e.g. for LIMIT).
func (w *Where) Next() int { return w.next }

type Sort struct {
	Column string
	Desc   bool
}

// ParseSort reads a comma-separated ?sort=field,-other spec against an
// allow-list mapping request field -> column. Unknown fields are skipped; an
// empty result falls back to the supplied default ordering.
func ParseSort(raw string, allowed map[string]string, fallback []Sort) []Sort {
	var out []Sort
	for tok := range strings.SplitSeq(raw, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		desc := false
		if strings.HasPrefix(tok, "-") {
			desc = true
			tok = tok[1:]
		}
		if col, ok := allowed[tok]; ok {
			out = append(out, Sort{Column: col, Desc: desc})
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

// OrderBy renders sorts as an ORDER BY body (without the keyword), or "".
func OrderBy(sorts []Sort) string {
	if len(sorts) == 0 {
		return ""
	}
	parts := make([]string, len(sorts))
	for i, s := range sorts {
		dir := "ASC"
		if s.Desc {
			dir = "DESC"
		}
		parts[i] = s.Column + " " + dir
	}
	return strings.Join(parts, ", ")
}
