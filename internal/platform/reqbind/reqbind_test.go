package reqbind

import "testing"

type validateNested struct {
	EngineSerial string `json:"engine_serial" binding:"omitempty,max=3"`
}

type validateSample struct {
	Name    string          `json:"name" binding:"required"`
	Vehicle *validateNested `json:"vehicle"`
}

func TestValidateNamesNestedFieldsByTheirJSONPath(t *testing.T) {
	got := Validate(&validateSample{Vehicle: &validateNested{EngineSerial: "TOO-LONG"}})
	if got["name"] != "this field is required" {
		t.Errorf(`name: got %q`, got["name"])
	}
	if got["vehicle.engine_serial"] != "must be at most 3 characters" {
		t.Errorf(`vehicle.engine_serial: got %q (all: %v)`, got["vehicle.engine_serial"], got)
	}
	if got := Validate(&validateSample{Name: "ok"}); got != nil {
		t.Errorf("valid value: got %v, want nil", got)
	}
}
