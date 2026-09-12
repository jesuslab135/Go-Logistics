package bulk

import (
	"encoding/json"
	"testing"
)

func TestCellValue(t *testing.T) {
	cases := []struct {
		col     Column
		cell    string
		want    string
		wantErr string
	}{
		{Column{Kind: KindString}, "", "", ""},
		{Column{Kind: KindString}, "Taller", `"Taller"`, ""},
		{Column{Kind: KindString, Max: 3}, "Taller", "", "must be at most 3 characters"},
		{Column{Kind: KindString, OneOf: []string{"truck", "trailer"}}, "Truck", `"truck"`, ""},
		{Column{Kind: KindString, OneOf: []string{"truck", "trailer"}}, "van", "", "must be one of: truck, trailer"},
		{Column{Kind: KindInt}, "12", "12", ""},
		{Column{Kind: KindInt}, "12.0", "12", ""},
		{Column{Kind: KindInt}, "12.5", "", "must be a whole number"},
		{Column{Kind: KindDecimal}, "$1,234.50", `"1234.5"`, ""},
		{Column{Kind: KindDecimal}, "doce", "", "must be a number"},
		{Column{Kind: KindBool}, "Sí", "true", ""},
		{Column{Kind: KindBool}, "si", "true", ""},
		{Column{Kind: KindBool}, "true", "true", ""},
		{Column{Kind: KindBool}, "yes", "true", ""},
		{Column{Kind: KindBool}, "1", "true", ""},
		{Column{Kind: KindBool}, "x", "true", ""},
		{Column{Kind: KindBool}, "verdadero", "true", ""},
		{Column{Kind: KindBool}, "no", "false", ""},
		{Column{Kind: KindBool}, "false", "false", ""},
		{Column{Kind: KindBool}, "0", "false", ""},
		{Column{Kind: KindBool}, "falso", "false", ""},
		{Column{Kind: KindBool}, "quizá", "", "must be true or false (sí/no)"},
		{Column{Kind: KindTime}, "2026-09-11", `"2026-09-11T00:00:00Z"`, ""},
		{Column{Kind: KindTime}, "2026-09-11 08:30", `"2026-09-11T08:30:00Z"`, ""},
		{Column{Kind: KindTime}, "2026-09-11 08:30:45", `"2026-09-11T08:30:45Z"`, ""},
		{Column{Kind: KindTime}, "11/09/2026", `"2026-09-11T00:00:00Z"`, ""},
		{Column{Kind: KindTime}, "11/09/2026 08:30", `"2026-09-11T08:30:00Z"`, ""},
		{Column{Kind: KindTime}, "2026-09-11T08:30:00-05:00", `"2026-09-11T13:30:00Z"`, ""},
		{Column{Kind: KindTime}, "46276", `"2026-09-11T00:00:00Z"`, ""},
		{Column{Kind: KindTime}, "ayer", "", "must be a date (YYYY-MM-DD or DD/MM/YYYY)"},
		{Column{Kind: KindJSON}, `{"a":1}`, `{"a":1}`, ""},
		{Column{Kind: KindJSON}, `{a:1}`, "", "must be valid JSON"},
		{Column{Kind: KindList}, "frenos, llantas; urgente", `["frenos","llantas","urgente"]`, ""},
		{Column{Kind: KindFloat}, "abc", "", "must be a number"},
		{Column{Kind: KindFloat}, "NaN", "", "must be a number"},
		{Column{Kind: KindFloat}, "Infinity", "", "must be a number"},
		{Column{Kind: KindInt}, "99999999999999999999999999999", "", "must be a whole number"},
		{Column{Kind: KindInt}, "Infinity", "", "must be a whole number"},
		{Column{Kind: KindFloat, CustomType: "number"}, "3.5", "3.5", ""},
		{Column{Kind: KindFloat, CustomType: "number"}, "no-es-numero", "", "must be a number"},
		// ParseFloat accepts these, but none is a valid JSON number, so the
		// raw cell used to reach Assemble and blow up there with no column
		// named. They are re-serialised instead.
		{Column{Kind: KindFloat}, ".5", "0.5", ""},
		{Column{Kind: KindFloat}, "5.", "5", ""},
		{Column{Kind: KindFloat}, "+5", "5", ""},
		{Column{Kind: KindFloat, CustomType: "number"}, ".5", "0.5", ""},
		{Column{Kind: KindString, CustomType: "date"}, "11/09/2026", `"2026-09-11"`, ""},
		{Column{Kind: KindString, CustomType: "select", OneOf: []string{"A12", "B7"}}, "a12", `"A12"`, ""},
	}
	for _, tc := range cases {
		got, err := CellValue(tc.col, tc.cell)
		if tc.wantErr != "" {
			if err == nil || err.Error() != tc.wantErr {
				t.Errorf("%+v %q: err = %v, want %q", tc.col, tc.cell, err, tc.wantErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("%+v %q: unexpected error %v", tc.col, tc.cell, err)
			continue
		}
		if string(got) != tc.want {
			t.Errorf("%+v %q: got %s, want %s", tc.col, tc.cell, got, tc.want)
		}
	}
}

// TestCellValueDecimalSeparators pins how a money cell is read when the person
// who filled the sheet in used Spanish conventions, English ones, or both at
// once. The old code stripped commas as thousands marks, so "1,5" became 15
// and "0,75" became 75 - a price wrong by 10x or 100x, committed with no
// error. Every case below is a real shape Excel produces.
func TestCellValueDecimalSeparators(t *testing.T) {
	cases := []struct {
		cell, want, wantErr string
	}{
		// The corruption this replaces: comma as a Spanish decimal separator.
		{"1,5", `"1.5"`, ""},
		{"0,75", `"0.75"`, ""},
		// Both separators present: the last one is the decimal separator.
		{"1,234.56", `"1234.56"`, ""},
		{"1.234,56", `"1234.56"`, ""},
		{"$1,234.50", `"1234.5"`, ""},
		{"1,234,567.89", `"1234567.89"`, ""},
		// One comma over an exact three-digit group is thousands.
		{"1,234", `"1234"`, ""},
		// Spaces group thousands in Spanish Excel too.
		{"1 234,56", `"1234.56"`, ""},
		// Repeated separators can only ever be grouping.
		{"1.234.567", `"1234567"`, ""},
		// Plain machine numbers - what this API's own exports emit - untouched.
		{"1234.56", `"1234.56"`, ""},
		{"1234", `"1234"`, ""},
		{"-1,5", `"-1.5"`, ""},
		// Refused rather than guessed at.
		{"1.234", "", errAmbiguous.Error()},
		{"1,23,456", "", errGrouping.Error()},
		{"doce", "", "must be a number"},
	}
	for _, tc := range cases {
		got, err := CellValue(Column{Kind: KindDecimal}, tc.cell)
		if tc.wantErr != "" {
			if err == nil || err.Error() != tc.wantErr {
				t.Errorf("%q: err = %v, want %q", tc.cell, err, tc.wantErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("%q: unexpected error %v", tc.cell, err)
			continue
		}
		if string(got) != tc.want {
			t.Errorf("%q: got %s, want %s", tc.cell, got, tc.want)
		}
	}
}

func TestAssembleNestsDottedKeysAndCustomFields(t *testing.T) {
	doc, err := Assemble(map[string]json.RawMessage{
		"name":                  json.RawMessage(`"Unidad 101"`),
		"vehicle.engine_serial": json.RawMessage(`"X1"`),
		"cf.cost_centre":        json.RawMessage(`"A12"`),
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(doc, &got); err != nil {
		t.Fatal(err)
	}
	if got["name"] != "Unidad 101" ||
		got["vehicle"].(map[string]any)["engine_serial"] != "X1" ||
		got["custom_fields"].(map[string]any)["cost_centre"] != "A12" {
		t.Fatalf("assembled %s", doc)
	}
}

// TestAssembleRejectsKeyCollision asserts the collision between the flat
// "vehicle" key and the dotted "vehicle.engine_serial" key is rejected
// regardless of which one Assemble happens to place into the tree first:
// child-before-parent (the nested object exists when the flat value is
// written) and parent-before-child (the flat value exists when the nested
// object is built). Go does not guarantee map iteration order, so the guard
// itself - not the order keys are listed in either literal below - must
// cover both directions.
func TestAssembleRejectsKeyCollision(t *testing.T) {
	if _, err := Assemble(map[string]json.RawMessage{
		"vehicle":               json.RawMessage(`"flat"`),
		"vehicle.engine_serial": json.RawMessage(`"X1"`),
	}); err == nil {
		t.Fatal("want an error when a parent key collides with one of its own nested children")
	}
	if _, err := Assemble(map[string]json.RawMessage{
		"vehicle.engine_serial": json.RawMessage(`"X1"`),
		"vehicle":               json.RawMessage(`"flat"`),
	}); err == nil {
		t.Fatal("want an error regardless of which colliding key is assembled first")
	}
}

func TestDecodeReportsColumns(t *testing.T) {
	_, errs := Decode[sampleRequest]([]byte(`{"vehicle":{"engine_serial":""}}`))
	if errs["name"] != "this field is required" {
		t.Fatalf("missing required field: got %v", errs)
	}
	_, errs = Decode[sampleRequest]([]byte(`{"name":"ok","count":"many"}`))
	if _, ok := errs["count"]; !ok {
		t.Fatalf("wrong type: got %v, want an error on count", errs)
	}
	in, errs := Decode[sampleRequest]([]byte(`{"name":"ok","count":3}`))
	if errs != nil || in.Name != "ok" || in.Count != 3 {
		t.Fatalf("valid row: got %+v, %v", in, errs)
	}
}
