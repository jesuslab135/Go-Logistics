package customfield

import (
	"encoding/json"
	"testing"
)

func defs() []Definition {
	return []Definition{
		{Key: "cost_centre", Label: "Cost centre", Type: TypeText, Required: true},
		{Key: "axles", Label: "Axles", Type: TypeNumber},
		{Key: "certified_until", Label: "Certified until", Type: TypeDate},
		{Key: "hazmat", Label: "Hazmat", Type: TypeBoolean},
		{Key: "region", Label: "Region", Type: TypeSelect, Options: []string{"north", "south"}},
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		doc      string
		wantKeys []string
	}{
		{
			name: "a document satisfying every definition",
			doc:  `{"cost_centre":"A12","axles":3,"certified_until":"2027-01-31","hazmat":true,"region":"north"}`,
		},
		{
			name:     "a missing required field",
			doc:      `{"axles":3}`,
			wantKeys: []string{"custom_fields.cost_centre"},
		},
		{
			name:     "an explicit null does not satisfy a required field",
			doc:      `{"cost_centre":null}`,
			wantKeys: []string{"custom_fields.cost_centre"},
		},
		{
			name:     "a number sent as a string",
			doc:      `{"cost_centre":"A12","axles":"3"}`,
			wantKeys: []string{"custom_fields.axles"},
		},
		{
			name:     "a timestamp where a business date belongs",
			doc:      `{"cost_centre":"A12","certified_until":"2027-01-31T00:00:00Z"}`,
			wantKeys: []string{"custom_fields.certified_until"},
		},
		{
			name:     "a select value outside its options",
			doc:      `{"cost_centre":"A12","region":"east"}`,
			wantKeys: []string{"custom_fields.region"},
		},
		{
			name:     "a boolean sent as a string",
			doc:      `{"cost_centre":"A12","hazmat":"yes"}`,
			wantKeys: []string{"custom_fields.hazmat"},
		},
		{
			// Undeclared keys are the tenant's own data. Rejecting them would
			// make records saved before a definition existed uneditable through
			// a form that no longer shows the field holding them.
			name: "an undeclared key is preserved, not rejected",
			doc:  `{"cost_centre":"A12","legacy_note":"kept"}`,
		},
		{
			name:     "several problems are all reported",
			doc:      `{"axles":"three","region":"east"}`,
			wantKeys: []string{"custom_fields.cost_centre", "custom_fields.axles", "custom_fields.region"},
		},
		{
			name:     "a document that is not an object",
			doc:      `["not","an","object"]`,
			wantKeys: []string{"custom_fields"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Validate(json.RawMessage(tt.doc), defs())

			if len(tt.wantKeys) == 0 {
				if got != nil {
					t.Fatalf("Validate() = %v, want no errors", got)
				}
				return
			}
			if len(got) != len(tt.wantKeys) {
				t.Fatalf("Validate() = %v, want errors on exactly %v", got, tt.wantKeys)
			}
			for _, key := range tt.wantKeys {
				if got[key] == "" {
					t.Errorf("no error reported for %s; got %v", key, got)
				}
			}
		})
	}
}

// An empty document against no definitions is the common case for the eight
// resources whose custom_fields nobody has declared anything for yet.
func TestValidateAcceptsAnEmptyDocumentWithNoDefinitions(t *testing.T) {
	if got := Validate(nil, nil); got != nil {
		t.Errorf("Validate(nil, nil) = %v, want no errors", got)
	}
}

// A required field is only required where somebody declared it required, so an
// unrelated resource's definitions must not leak in.
func TestValidateRequiredOnlyForDeclaredFields(t *testing.T) {
	got := Validate(json.RawMessage(`{}`), []Definition{
		{Key: "optional", Label: "Optional", Type: TypeText},
	})
	if got != nil {
		t.Errorf("Validate() = %v, want no errors for an absent optional field", got)
	}
}

// A filter value has to be the JSON scalar the stored document holds, or
// containment silently matches nothing: "3" is not 3.
func TestFilterValue(t *testing.T) {
	tests := []struct {
		name    string
		def     Definition
		raw     string
		want    string
		wantErr bool
	}{
		{name: "text is quoted", def: Definition{Type: TypeText}, raw: "A12", want: `"A12"`},
		{name: "a select value is quoted", def: Definition{Type: TypeSelect}, raw: "north", want: `"north"`},
		{name: "a date is quoted", def: Definition{Type: TypeDate}, raw: "2027-01-31", want: `"2027-01-31"`},
		{name: "a number is bare", def: Definition{Type: TypeNumber}, raw: "3", want: `3`},
		{name: "a boolean is bare", def: Definition{Type: TypeBoolean}, raw: "true", want: `true`},
		{name: "a non-numeric number is refused", def: Definition{Type: TypeNumber}, raw: "three", wantErr: true},
		{name: "a non-boolean boolean is refused", def: Definition{Type: TypeBoolean}, raw: "yes", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilterValue(tt.def, tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("FilterValue() = %s, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("FilterValue() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("FilterValue() = %s, want %s", got, tt.want)
			}
		})
	}
}
