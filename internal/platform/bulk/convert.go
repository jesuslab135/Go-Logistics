package bulk

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shopspring/decimal"

	"fleet/internal/platform/reqbind"
)

var excelEpoch = time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)

var dateLayouts = []string{
	"2006-01-02", "2006-01-02 15:04", "2006-01-02 15:04:05",
	"02/01/2006", "02/01/2006 15:04", time.RFC3339,
}

var (
	errWhole  = errors.New("must be a whole number")
	errNumber = errors.New("must be a number")
	errBool   = errors.New("must be true or false (sí/no)")
	errDate   = errors.New("must be a date (YYYY-MM-DD or DD/MM/YYYY)")
	errJSON   = errors.New("must be valid JSON")

	// errGrouping and errAmbiguous are the two ways a number can be refused
	// for its separators rather than its digits. Both exist so a doubtful cell
	// is rejected by name instead of being silently turned into the wrong
	// amount: a refused row can be fixed and re-uploaded, a wrong price that
	// was committed cannot be found again.
	errGrouping  = errors.New("must be a number: thousands separators must group digits in threes (1,234.56 or 1.234,56)")
	errAmbiguous = errors.New("is ambiguous: a comma before exactly three digits could be a thousands separator or a decimal separator; write 1234 or 1,234.00 if it is thousands, or 1.234 if it is a decimal")
)

// CellValue turns one cell into the JSON value its column expects. An empty
// cell returns nil, nil: the key is left out, so the store's defaults apply.
func CellValue(col Column, cell string) (json.RawMessage, error) {
	cell = strings.TrimSpace(cell)
	if cell == "" {
		return nil, nil
	}
	switch col.CustomType {
	case "number":
		col.Kind = KindFloat
	case "boolean":
		col.Kind = KindBool
	case "date":
		t, err := parseTime(cell)
		if err != nil {
			return nil, err
		}
		return json.Marshal(t.Format(time.DateOnly))
	}

	switch col.Kind {
	case KindInt:
		n, err := strconv.ParseInt(cell, 10, 64)
		if err != nil {
			f, ferr := strconv.ParseFloat(cell, 64)
			if ferr != nil || math.IsNaN(f) || math.IsInf(f, 0) ||
				f < float64(math.MinInt64) || f >= float64(math.MaxInt64) ||
				f != math.Trunc(f) {
				return nil, errWhole
			}
			n = int64(f)
		}
		return json.RawMessage(strconv.FormatInt(n, 10)), nil
	case KindDecimal:
		normal, err := normalizeDecimal(cell)
		if err != nil {
			return nil, err
		}
		d, err := decimal.NewFromString(normal)
		if err != nil {
			return nil, errNumber
		}
		return json.Marshal(d.String())
	case KindFloat:
		f, err := strconv.ParseFloat(cell, 64)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, errNumber
		}
		// Re-serialise rather than echoing the cell: ParseFloat accepts ".5",
		// "5." and "+5", none of which are valid JSON numbers, so passing the
		// raw text through would leave Assemble's json.Marshal to fail later
		// with a raw Go error naming no column at all. KindInt already does
		// this with FormatInt; every custom field of type "number" lands here.
		return json.RawMessage(strconv.FormatFloat(f, 'f', -1, 64)), nil
	case KindBool:
		switch strings.ToLower(cell) {
		case "true", "verdadero", "sí", "si", "yes", "1", "x":
			return json.RawMessage("true"), nil
		case "false", "falso", "no", "0":
			return json.RawMessage("false"), nil
		}
		return nil, errBool
	case KindTime:
		t, err := parseTime(cell)
		if err != nil {
			return nil, err
		}
		return json.Marshal(t.UTC().Format(time.RFC3339))
	case KindJSON:
		if !json.Valid([]byte(cell)) {
			return nil, errJSON
		}
		return json.RawMessage(cell), nil
	case KindList:
		parts := strings.FieldsFunc(cell, func(r rune) bool { return r == ',' || r == ';' })
		list := make([]string, 0, len(parts))
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				list = append(list, p)
			}
		}
		return json.Marshal(list)
	default: // KindString
		if len(col.OneOf) > 0 {
			for _, v := range col.OneOf {
				if strings.EqualFold(v, cell) {
					return json.Marshal(v)
				}
			}
			return nil, fmt.Errorf("must be one of: %s", strings.Join(col.OneOf, ", "))
		}
		if col.Max > 0 && utf8.RuneCountInString(cell) > col.Max {
			return nil, fmt.Errorf("must be at most %d characters", col.Max)
		}
		return json.Marshal(cell)
	}
}

func parseTime(cell string) (time.Time, error) {
	if serial, err := strconv.ParseFloat(cell, 64); err == nil && serial > 0 {
		return excelEpoch.Add(time.Duration(math.Round(serial*86400)) * time.Second), nil
	}
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, cell); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errDate
}

// decimalStripper removes the decoration a spreadsheet puts around an amount:
// a currency mark and the ordinary, non-breaking and narrow spaces Excel uses
// to group thousands.
var decimalStripper = strings.NewReplacer("$", "", " ", "", " ", "", " ", "")

// normalizeDecimal rewrites a human-written amount as a plain machine number,
// working out which separator means what instead of assuming.
//
// This endpoint is deliberately aimed at Spanish users - sheet.readCSV sniffs
// the ';' delimiter because Spanish Excel writes it, the templates are in
// Spanish, and booleans accept "sí" - and in Spanish the comma is the DECIMAL
// separator. Stripping commas as thousands marks (what this used to do) turned
// "1,5" into 15 and "0,75" into 75, committing an amount off by 10x or 100x
// with no error at all. The rules are:
//
//   - "$" and spaces are stripped.
//   - With BOTH separators present the LAST one is the decimal separator and
//     the other groups thousands, so "1,234.56" and "1.234,56" are both
//     1234.56.
//   - With only one separator character present, several occurrences can only
//     be thousands grouping: "1.234.567" and "1,234,567" are both 1234567.
//   - A lone PERIOD is always the decimal point: "1.234" is 1.234 and
//     "19.432" is 19.432. Excel stores every number in period-decimal form
//     whatever the user's display locale, and readXLSX reads raw cell values,
//     so an .xlsx number arrives in exactly this shape with nothing ambiguous
//     about it. KindDecimal is every decimal.Decimal field (see schema.go),
//     not just money: latitude, longitude, odometer, meter readings,
//     quantities and annual_percentage_rate all land here.
//   - A lone COMMA before exactly three digits is the one genuinely
//     undecidable shape: "1,234" is 1234 to an English writer and 1.234 to a
//     Spanish one, and nothing in the cell says which. It is refused
//     (errAmbiguous) rather than guessed at - silently taking the English
//     reading would be the same 1000x corruption this function exists to
//     prevent, just moved from "1,5" to "1,234".
//   - Any other lone comma is the decimal point: "1,5" is 1.5, "0,75" is 0.75.
//   - Grouping that is claimed but malformed ("1,23,456") is refused too.
func normalizeDecimal(cell string) (string, error) {
	s := decimalStripper.Replace(cell)
	lastDot := strings.LastIndexByte(s, '.')
	lastComma := strings.LastIndexByte(s, ',')

	switch {
	case lastDot >= 0 && lastComma >= 0:
		if lastComma > lastDot {
			return regroup(s, ',', '.')
		}
		return regroup(s, '.', ',')
	case lastComma >= 0:
		return oneSeparator(s, ',')
	case lastDot >= 0:
		return oneSeparator(s, '.')
	}
	return s, nil
}

// oneSeparator resolves a number written with only sep in it.
func oneSeparator(s string, sep byte) (string, error) {
	if strings.Count(s, string(sep)) > 1 {
		// Repeated, it can only ever be thousands grouping: two decimal
		// points in one number is not a reading anybody intends.
		return regroup(s, 0, sep)
	}
	// The comma is the only separator whose single occurrence can be either
	// thing. A period is unambiguous by construction - see normalizeDecimal -
	// and treating it as grouping refused every .xlsx latitude, odometer and
	// quantity, which arrive as raw period-decimal values.
	if sep == ',' && isThousandsGroup(s, sep) {
		return "", errAmbiguous
	}
	return strings.Replace(s, string(sep), ".", 1), nil
}

// isThousandsGroup reports whether s is exactly one group boundary: an
// optionally signed head of 1-3 digits, sep, then exactly 3 digits ("1,234").
func isThousandsGroup(s string, sep byte) bool {
	head, tail, ok := strings.Cut(s, string(sep))
	if !ok || len(tail) != 3 || !allDigits(tail) {
		return false
	}
	head = strings.TrimPrefix(strings.TrimPrefix(head, "+"), "-")
	return len(head) <= 3 && allDigits(head)
}

// regroup drops the group separators once every group really is three digits
// and rewrites dec, when there is one, as the period decimal.NewFromString
// expects. A dec of 0 means the number has no fractional part.
func regroup(s string, dec, group byte) (string, error) {
	whole, frac := s, ""
	if dec != 0 {
		i := strings.LastIndexByte(s, dec)
		whole, frac = s[:i], s[i+1:]
		if !allDigits(frac) {
			return "", errNumber
		}
	}

	parts := strings.Split(whole, string(group))
	sign := ""
	if head := parts[0]; strings.HasPrefix(head, "+") || strings.HasPrefix(head, "-") {
		sign, parts[0] = head[:1], head[1:]
	}
	for i, p := range parts {
		if !allDigits(p) {
			return "", errNumber
		}
		// The leading group may be short ("1,234"); every later one is a full
		// group of three or the separators are not grouping at all.
		if (i == 0 && len(p) > 3) || (i > 0 && len(p) != 3) {
			return "", errGrouping
		}
	}

	out := sign + strings.Join(parts, "")
	if dec != 0 {
		out += "." + frac
	}
	return out, nil
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != ""
}

// Assemble builds the JSON object for one row: dotted keys become nested
// objects and cf.<key> becomes custom_fields.<key>. It returns an error if
// one key's path collides with another key already placed in the same row
// (for example both "vehicle" and "vehicle.engine_serial" present at once) -
// callers must not feed it a value map with such an overlap.
func Assemble(values map[string]json.RawMessage) ([]byte, error) {
	root := map[string]any{}
	for key, v := range values {
		fullKey := key
		if strings.HasPrefix(key, "cf.") {
			key = "custom_fields." + strings.TrimPrefix(key, "cf.")
		}
		parts := strings.Split(key, ".")
		node := root
		for _, p := range parts[:len(parts)-1] {
			switch existing := node[p].(type) {
			case nil:
				child := map[string]any{}
				node[p] = child
				node = child
			case map[string]any:
				node = existing
			default:
				return nil, fmt.Errorf("bulk: column %q conflicts with another column at %q", fullKey, p)
			}
		}
		last := parts[len(parts)-1]
		if _, isMap := node[last].(map[string]any); isMap {
			return nil, fmt.Errorf("bulk: column %q conflicts with another column at %q", fullKey, last)
		}
		node[last] = v
	}
	return json.Marshal(root)
}

// Decode fills a create request from one assembled row and validates it with
// the rules the JSON API applies. The map has one message per offending
// column; it is nil when the row is valid.
func Decode[C any](doc []byte) (C, map[string]string) {
	var in C
	if err := json.Unmarshal(doc, &in); err != nil {
		var te *json.UnmarshalTypeError
		if errors.As(err, &te) && te.Field != "" {
			return in, map[string]string{te.Field: "has the wrong type of value"}
		}
		return in, map[string]string{"": "the row could not be read: " + err.Error()}
	}
	return in, reqbind.Validate(&in)
}
