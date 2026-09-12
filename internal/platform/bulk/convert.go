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
		d, err := decimal.NewFromString(strings.NewReplacer("$", "", ",", "", " ", "").Replace(cell))
		if err != nil {
			return nil, errNumber
		}
		return json.Marshal(d.String())
	case KindFloat:
		f, err := strconv.ParseFloat(cell, 64)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, errNumber
		}
		return json.RawMessage(cell), nil
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
