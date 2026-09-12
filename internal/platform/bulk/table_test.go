package bulk

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTableFlattensInFirstSeenOrder(t *testing.T) {
	var tb Table
	for _, raw := range []string{
		`{"id":1,"name":"Unidad 101","vehicle":{"engine_serial":"X1"},"labels":["a","b"],"cost":"12.50","active":true,"notes":null}`,
		`{"id":2,"name":"Unidad 102","extra":7}`,
	} {
		if err := tb.Add(json.RawMessage(raw)); err != nil {
			t.Fatal(err)
		}
	}
	want := [][]any{
		{"id", "name", "vehicle.engine_serial", "labels", "cost", "active", "notes", "extra"},
		{int64(1), "Unidad 101", "X1", `["a","b"]`, "12.50", true, nil, nil},
		{int64(2), "Unidad 102", nil, nil, nil, nil, nil, int64(7)},
	}
	if got := tb.Records(); !reflect.DeepEqual(got, want) {
		t.Fatalf("records:\n got %#v\nwant %#v", got, want)
	}
}

// The import engine keys a custom field's column as "cf.<key>" (engine.go's
// Flat.Columns), not "custom_fields.<key>". A list response nests custom
// fields under "custom_fields", so Table must remap that one prefix or a
// file exported from any section with custom fields would come back with
// headers the importer silently ignores, losing every custom-field value on
// re-upload.
func TestTableRemapsCustomFieldsToCfPrefix(t *testing.T) {
	var tb Table
	if err := tb.Add(json.RawMessage(`{"id":1,"custom_fields":{"color":"red","axles":2}}`)); err != nil {
		t.Fatal(err)
	}
	want := [][]any{
		{"id", "cf.color", "cf.axles"},
		{int64(1), "red", int64(2)},
	}
	if got := tb.Records(); !reflect.DeepEqual(got, want) {
		t.Fatalf("records:\n got %#v\nwant %#v", got, want)
	}
}

// A nested object that is populated on one row and null on another must not
// register both a dotted column and a bare one: which form shows up first
// would depend on row order, so two exports of the same section could
// disagree on their own header row.
func TestTableDropsBareColumnShadowedByNestedObject(t *testing.T) {
	var tb Table
	for _, raw := range []string{
		`{"id":1,"vehicle":{"engine_serial":"X1"}}`,
		`{"id":2,"vehicle":null}`,
	} {
		if err := tb.Add(json.RawMessage(raw)); err != nil {
			t.Fatal(err)
		}
	}
	want := [][]any{
		{"id", "vehicle.engine_serial"},
		{int64(1), "X1"},
		{int64(2), nil},
	}
	if got := tb.Records(); !reflect.DeepEqual(got, want) {
		t.Fatalf("records:\n got %#v\nwant %#v", got, want)
	}
}
