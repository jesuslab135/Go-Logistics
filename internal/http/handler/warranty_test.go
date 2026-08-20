package handler

import (
	"testing"
	"time"

	"fleet/internal/db/gen"
)

// VehicleDossier renders the provider as "#<id>" without this: the response
// carried provider_id but never the resolved name.
func TestToWarrantyResponseCarriesProviderName(t *testing.T) {
	now := time.Now().UTC()
	row := gen.ListWarrantiesRow{
		ID:           3,
		CompanyID:    1,
		ProviderID:   9,
		StartDate:    now,
		EndDate:      now,
		Terms:        "12 months",
		IsActive:     true,
		ProviderName: "Michelin",
	}

	got := toWarrantyResponse(row)
	if got.ProviderID != 9 {
		t.Errorf("provider_id = %d, want 9", got.ProviderID)
	}
	if got.ProviderName != "Michelin" {
		t.Errorf("provider_name = %q, want %q", got.ProviderName, "Michelin")
	}
}
