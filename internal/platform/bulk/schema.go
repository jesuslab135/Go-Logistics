// Package bulk imports spreadsheet rows through a section's own create
// request and store, builds the matching Excel template, and flattens exported
// lists into rows.
package bulk

import (
	"encoding/json"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Kind is how a column's cells are converted.
type Kind int

const (
	KindString Kind = iota
	KindInt
	KindDecimal
	KindFloat
	KindBool
	KindTime
	KindJSON
	KindList
)

// Column is one spreadsheet column.
type Column struct {
	// Key is the header: the API field name, dotted for a nested request
	// ("vehicle.engine_serial"), "cf.<key>" for a custom field.
	Key      string
	Kind     Kind
	Required bool
	Max      int      // maximum length of a string, 0 when unset
	OneOf    []string // permitted values
	Label    string   // human description for the Instrucciones sheet
	// Ref names the lookup a reference column resolves through (its own key).
	Ref string
	// CustomType is the custom field's type ("text", "number", "date",
	// "boolean", "select") for cf.* columns, empty otherwise.
	CustomType string
}

var (
	timeType    = reflect.TypeOf(time.Time{})
	decimalType = reflect.TypeOf(decimal.Decimal{})
	rawType     = reflect.TypeOf(json.RawMessage{})
)

// SchemaOf derives the columns of a create request type. JSON keys in skip,
// "-" fields and custom_fields are left out.
func SchemaOf(t reflect.Type, skip ...string) []Column {
	return schemaOf(t, "", skip)
}

func schemaOf(t reflect.Type, prefix string, skip []string) []Column {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	var cols []Column
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" || name == "custom_fields" {
			continue
		}
		key := prefix + name
		if slices.Contains(skip, key) {
			continue
		}

		ft := f.Type
		if ft.Kind() == reflect.Pointer && ft.Elem().Kind() == reflect.Struct &&
			ft.Elem() != timeType && ft.Elem() != decimalType {
			cols = append(cols, schemaOf(ft.Elem(), key+".", skip)...)
			continue
		}

		col := Column{Key: key, Kind: kindOf(ft, name)}
		for _, rule := range strings.Split(f.Tag.Get("binding"), ",") {
			switch {
			case rule == "required":
				col.Required = true
			case strings.HasPrefix(rule, "max="):
				col.Max, _ = strconv.Atoi(strings.TrimPrefix(rule, "max="))
			case strings.HasPrefix(rule, "oneof="):
				col.OneOf = strings.Fields(strings.TrimPrefix(rule, "oneof="))
			}
		}
		cols = append(cols, col)
	}
	return cols
}

func kindOf(t reflect.Type, name string) Kind {
	if t == rawType {
		if name == "labels" {
			return KindList
		}
		return KindJSON
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch {
	case t == timeType:
		return KindTime
	case t == decimalType:
		return KindDecimal
	}
	switch t.Kind() {
	case reflect.String:
		return KindString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return KindInt
	case reflect.Float32, reflect.Float64:
		return KindFloat
	case reflect.Bool:
		return KindBool
	case reflect.Slice:
		if t.Elem().Kind() == reflect.String {
			return KindList
		}
	}
	return KindJSON
}
