package handler

import (
	"encoding/json"
	"testing"

	"fleet/internal/http/dto"
)

func TestCreateAssetRequestBindsNestedExtension(t *testing.T) {
	body := `{"name":"Dry Van #1","vin_sn":"X1","trailer":{"trailer_type":"DRY_VAN"}}`
	var in dto.CreateAssetRequest
	if err := json.Unmarshal([]byte(body), &in); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if in.Trailer == nil || in.Trailer.TrailerType != "DRY_VAN" {
		t.Fatalf("trailer not bound: %+v", in.Trailer)
	}
	if in.Vehicle != nil {
		t.Error("vehicle should be nil when omitted")
	}
}

func TestAssetOnlyRequestHasNoExtension(t *testing.T) {
	var in dto.CreateAssetRequest
	_ = json.Unmarshal([]byte(`{"name":"Forklift","vin_sn":"F1"}`), &in)
	if in.Vehicle != nil || in.Trailer != nil {
		t.Error("asset-only request must carry no extension")
	}
}
