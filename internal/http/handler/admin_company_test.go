package handler

import (
	"encoding/json"
	"testing"

	"fleet/internal/db/gen"
)

// The admin company row is the tenant CompanyResponse plus the client it
// belongs to. Embedding must flatten into one JSON object: a nested
// "CompanyResponse" key would break every field the frontend already reads.
func TestToAdminCompanyResponseFlattensTenantAndAccountFields(t *testing.T) {
	logo := "logos/alpha.png"
	got := toAdminCompanyResponse(gen.Company{
		ID: 7, Name: "Alpha Fleet", TaxID: "ALPHA-1", Logo: &logo, AccountID: 3,
	}, "Cliente Alpha", 4)

	body, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var flat map[string]any
	if err := json.Unmarshal(body, &flat); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	want := map[string]any{
		"id":             float64(7),
		"name":           "Alpha Fleet",
		"tax_id":         "ALPHA-1",
		"logo":           "logos/alpha.png",
		"account_id":     float64(3),
		"account_name":   "Cliente Alpha",
		"employee_count": float64(4),
	}
	for k, v := range want {
		if flat[k] != v {
			t.Errorf("%s = %v, want %v (body %s)", k, flat[k], v, body)
		}
	}
	if _, nested := flat["CompanyResponse"]; nested {
		t.Fatalf("CompanyResponse must be flattened, got %s", body)
	}
}
