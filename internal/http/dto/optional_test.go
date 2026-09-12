package dto

import (
	"encoding/json"
	"testing"
)

// The whole point of OptionalInt64: an omitted field and an explicit null must
// decode differently, or a client that never sends role_id would clear it.
func TestOptionalInt64TellsOmittedFromNull(t *testing.T) {
	type body struct {
		RoleID OptionalInt64 `json:"role_id"`
	}

	cases := []struct {
		name    string
		in      string
		wantSet bool
		want    *int64
	}{
		{name: "omitted", in: `{}`, wantSet: false, want: nil},
		{name: "null", in: `{"role_id":null}`, wantSet: true, want: nil},
		{name: "number", in: `{"role_id":7}`, wantSet: true, want: ptrInt64(7)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var b body
			if err := json.Unmarshal([]byte(tc.in), &b); err != nil {
				t.Fatalf("unmarshal %s: %v", tc.in, err)
			}
			if b.RoleID.Set != tc.wantSet {
				t.Fatalf("Set = %v, want %v", b.RoleID.Set, tc.wantSet)
			}
			switch {
			case tc.want == nil && b.RoleID.Value != nil:
				t.Fatalf("Value = %d, want nil", *b.RoleID.Value)
			case tc.want != nil && (b.RoleID.Value == nil || *b.RoleID.Value != *tc.want):
				t.Fatalf("Value = %v, want %d", b.RoleID.Value, *tc.want)
			}
		})
	}

	var b body
	if err := json.Unmarshal([]byte(`{"role_id":"x"}`), &b); err == nil {
		t.Fatal("a non-numeric role_id decoded without error")
	}
}

func ptrInt64(v int64) *int64 { return &v }
