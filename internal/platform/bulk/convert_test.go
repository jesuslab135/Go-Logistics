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
		{Column{Kind: KindBool}, "no", "false", ""},
		{Column{Kind: KindBool}, "quizá", "", "must be true or false (sí/no)"},
		{Column{Kind: KindTime}, "2026-09-11", `"2026-09-11T00:00:00Z"`, ""},
		{Column{Kind: KindTime}, "11/09/2026", `"2026-09-11T00:00:00Z"`, ""},
		{Column{Kind: KindTime}, "46276", `"2026-09-11T00:00:00Z"`, ""},
		{Column{Kind: KindTime}, "ayer", "", "must be a date (YYYY-MM-DD or DD/MM/YYYY)"},
		{Column{Kind: KindJSON}, `{"a":1}`, `{"a":1}`, ""},
		{Column{Kind: KindJSON}, `{a:1}`, "", "must be valid JSON"},
		{Column{Kind: KindList}, "frenos, llantas; urgente", `["frenos","llantas","urgente"]`, ""},
		{Column{Kind: KindFloat, CustomType: "number"}, "3.5", "3.5", ""},
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
