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
