package handler

import "testing"

// A key becomes a JSON key, a query parameter and a form field name all at
// once, so it has to be safe as all three.
func TestValidCustomFieldKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"cost_centre", true},
		{"axles2", true},
		{"a", true},
		{"", false},
		{"2axles", false},      // a leading digit is not a safe identifier
		{"Cost_Centre", false}, // case would make two keys look like one
		{"cost-centre", false}, // a hyphen collides with the query syntax
		{"cost centre", false}, // a space cannot survive a query string
		{"cost.centre", false}, // a dot is how the filter separates prefix from key
		{"_leading", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := validCustomFieldKey(tt.key); got != tt.want {
				t.Errorf("validCustomFieldKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

// A definition that describes a form nothing renders, or a select with nothing
// to select, is a fact about nothing.
func TestValidateDefinition(t *testing.T) {
	tests := []struct {
		name      string
		resource  string
		key       string
		fieldType string
		options   []string
		wantErr   bool
	}{
		{name: "a valid text field", resource: "assets", key: "cost_centre", fieldType: "text"},
		{name: "a valid select", resource: "assets", key: "region", fieldType: "select", options: []string{"north"}},
		{name: "a resource that carries no custom_fields", resource: "tires", key: "x", fieldType: "text", wantErr: true},
		{name: "an unknown field type", resource: "assets", key: "x", fieldType: "colour", wantErr: true},
		{name: "a select with no options", resource: "assets", key: "x", fieldType: "select", wantErr: true},
		{name: "options on something that is not a select", resource: "assets", key: "x", fieldType: "text", options: []string{"a"}, wantErr: true},
		{name: "an unusable key", resource: "assets", key: "Cost Centre", fieldType: "text", wantErr: true},
		// An update passes no resource or key, because neither may change.
		{name: "an update omits resource and key", resource: "", key: "", fieldType: "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDefinition(tt.resource, tt.key, tt.fieldType, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDefinition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
