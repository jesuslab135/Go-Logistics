package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"fleet/internal/http/dto"
)

// A create request that omits a field must not overwrite the column default with
// Go's zero value. employee.is_active is the case that bit: an employee created
// without it was inserted inactive and could never log in.
func TestCreateEmployeeDefaultsToActive(t *testing.T) {
	var in dto.CreateEmployeeRequest
	if err := json.Unmarshal([]byte(`{"first_name":"A","last_name":"B","email":"a@b.c"}`), &in); err != nil {
		t.Fatal(err)
	}
	if in.IsActive != nil {
		t.Fatalf("omitted is_active should decode to nil, got %v", *in.IsActive)
	}
	if got := createEmployeeParams(in, time.Now()); !got.IsActive {
		t.Error("employee created without is_active must be active")
	}
}

// An explicit false must still win, which is the whole reason the field is a
// pointer rather than a plain bool.
func TestCreateEmployeeHonoursExplicitFalse(t *testing.T) {
	var in dto.CreateEmployeeRequest
	if err := json.Unmarshal([]byte(`{"is_active":false}`), &in); err != nil {
		t.Fatal(err)
	}
	if got := createEmployeeParams(in, time.Now()); got.IsActive {
		t.Error(`"is_active":false was ignored`)
	}
}

func TestDefaultHelpers(t *testing.T) {
	yes, no := true, false

	if got := orDefault("", "OPEN"); got != "OPEN" {
		t.Errorf("orDefault(\"\") = %q, want OPEN", got)
	}
	if got := orDefault("CLOSED", "OPEN"); got != "CLOSED" {
		t.Errorf("orDefault kept the wrong value: %q", got)
	}
	if !boolOrDefault(nil, true) {
		t.Error("boolOrDefault(nil, true) must fall back to true")
	}
	if boolOrDefault(&no, true) {
		t.Error("boolOrDefault must honour an explicit false")
	}
	if !boolOrDefault(&yes, false) {
		t.Error("boolOrDefault must honour an explicit true")
	}
	if got := int32OrDefault(0, 1); got != 1 {
		t.Errorf("int32OrDefault(0, 1) = %d, want 1", got)
	}
	if got := int32OrDefault(7, 1); got != 7 {
		t.Errorf("int32OrDefault(7, 1) = %d, want 7", got)
	}
	if got := decimalOrDefault(decimal.Decimal{}, 1); !got.Equal(decimal.NewFromInt(1)) {
		t.Errorf("decimalOrDefault(zero, 1) = %s, want 1", got)
	}
	if got := decimalOrDefault(decimal.NewFromInt(4), 1); !got.Equal(decimal.NewFromInt(4)) {
		t.Errorf("decimalOrDefault kept the wrong value: %s", got)
	}
}

// ancestry is derived from parent_id in SQL, so the request DTO must not carry
// it — a client that could set both could describe a tree two ways at once.
func TestCreateGroupRequestRejectsAncestry(t *testing.T) {
	var in dto.CreateGroupRequest
	if err := json.Unmarshal([]byte(`{"name":"Norte","ancestry":"9/9"}`), &in); err != nil {
		t.Fatal(err)
	}
	if in.Name != "Norte" {
		t.Fatalf("name did not decode: %q", in.Name)
	}
	// The struct simply has no such field; this is a compile-time guarantee that
	// the test pins by round-tripping the type.
	out, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var back map[string]any
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatal(err)
	}
	if _, ok := back["ancestry"]; ok {
		t.Error("CreateGroupRequest must not carry ancestry")
	}
}
