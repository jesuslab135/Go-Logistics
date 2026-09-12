//go:build integration

package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// The nested line-item deletes are tenant-scoped in SQL, but they were declared
// :exec — so a DELETE aimed at another company's document matched zero rows and
// returned nil, and the handler went on to recompute the document it had just
// failed to touch. recalc* carried no tenant, so the recompute landed on the
// victim's row: an attacker's 204 rewrote another company's totals and, through
// Clear*TotalOverride, destroyed their audited override, its reason and its
// approver.
//
// Zero rows must be a 404, and the recompute must never run for a document the
// caller was not allowed to write.
func TestDeletingAnotherCompanysLineItemIs404(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	srv := httptest.NewServer(newIntegrationRouter(pool))
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	_, _, _, ownerToken := provisionAccountOwner(t, srv.URL, platformToken)

	ts := time.Now().UnixNano()
	companyA := createCompany(t, srv.URL, ownerToken, "Company A", fmt.Sprintf("TAX-XT-A-%d", ts))
	companyB := createCompany(t, srv.URL, ownerToken, "Company B", fmt.Sprintf("TAX-XT-B-%d", ts))
	tokenA := switchCompany(t, srv.URL, ownerToken, companyA)
	tokenB := switchCompany(t, srv.URL, ownerToken, companyB)

	victim := seedWorkOrder(t, srv.URL, tokenB, ts, "B")

	// Company B records a total override: a deliberate, audited figure that a
	// recompute would clear.
	postJSON(t, srv.URL+"/api/v1/work-orders/"+itoa(victim)+"/override-total", tokenB, map[string]any{
		"amount": "4242.00",
		"reason": "agreed with the vendor",
	}, http.StatusOK, nil)

	before := workOrderOverride(t, srv.URL, tokenB, victim)
	if before == nil {
		t.Fatal("precondition failed: company B's total override was not recorded")
	}

	// Company A aims a delete at company B's work order. The line item id does
	// not matter — nothing under that work order belongs to A.
	for _, path := range []string{
		"/api/v1/work-orders/" + itoa(victim) + "/line-items/1",
		"/api/v1/work-orders/" + itoa(victim) + "/line-items/1/sub-line-items/1",
	} {
		t.Run(path, func(t *testing.T) {
			doJSON(t, http.MethodDelete, srv.URL+path, tokenA, nil, http.StatusNotFound, nil)
		})
	}

	if after := workOrderOverride(t, srv.URL, tokenB, victim); after == nil || *after != *before {
		t.Errorf("company B's total override changed from %v to %v after company A's delete",
			deref(before), deref(after))
	}
}

// Purchase orders and service entries carry the same delete-then-recompute
// shape, so they carry the same defect. Proven separately rather than assumed
// from the work-order case.
func TestDeletingAnotherCompanysPurchaseOrderLineIs404(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	srv := httptest.NewServer(newIntegrationRouter(pool))
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	_, _, _, ownerToken := provisionAccountOwner(t, srv.URL, platformToken)

	ts := time.Now().UnixNano()
	companyA := createCompany(t, srv.URL, ownerToken, "Company A", fmt.Sprintf("TAX-XP-A-%d", ts))
	companyB := createCompany(t, srv.URL, ownerToken, "Company B", fmt.Sprintf("TAX-XP-B-%d", ts))
	tokenA := switchCompany(t, srv.URL, ownerToken, companyA)
	tokenB := switchCompany(t, srv.URL, ownerToken, companyB)

	// purchase_order.vendor_id and destination_id are both NOT NULL, so B needs
	// a vendor and a location before it can own a purchase order.
	var vendorB struct {
		ID int64 `json:"id"`
	}
	postJSON(t, srv.URL+"/api/v1/vendors", tokenB, map[string]any{
		"name": fmt.Sprintf("Proveedor B %d", ts),
	}, http.StatusCreated, &vendorB)

	// destination_id references part_location, not location.
	var locationB struct {
		ID int64 `json:"id"`
	}
	postJSON(t, srv.URL+"/api/v1/part-locations", tokenB, map[string]any{
		"name": fmt.Sprintf("Almacén B %d", ts),
	}, http.StatusCreated, &locationB)

	var poB struct {
		ID int64 `json:"id"`
	}
	postJSON(t, srv.URL+"/api/v1/purchase-orders", tokenB, map[string]any{
		"number":         fmt.Sprintf("PO-B-%d", ts),
		"vendor_id":      vendorB.ID,
		"destination_id": locationB.ID,
	}, http.StatusCreated, &poB)

	doJSON(t, http.MethodDelete,
		srv.URL+"/api/v1/purchase-orders/"+itoa(poB.ID)+"/line-items/1",
		tokenA, nil, http.StatusNotFound, nil)
}

// The recompute family reads and writes the document directly. It used to do so
// without a tenant, on the argument that the triggering write had already
// established one. This aims a recompute trigger — the audited total override —
// at another company's work order, and requires that the document is neither
// read nor rewritten.
func TestOverridingAnotherCompanysTotalIsRefused(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	srv := httptest.NewServer(newIntegrationRouter(pool))
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	_, _, _, ownerToken := provisionAccountOwner(t, srv.URL, platformToken)

	ts := time.Now().UnixNano()
	companyA := createCompany(t, srv.URL, ownerToken, "Company A", fmt.Sprintf("TAX-OV-A-%d", ts))
	companyB := createCompany(t, srv.URL, ownerToken, "Company B", fmt.Sprintf("TAX-OV-B-%d", ts))
	tokenA := switchCompany(t, srv.URL, ownerToken, companyA)
	tokenB := switchCompany(t, srv.URL, ownerToken, companyB)

	victim := seedWorkOrder(t, srv.URL, tokenB, ts, "B")
	postJSON(t, srv.URL+"/api/v1/work-orders/"+itoa(victim)+"/override-total", tokenB, map[string]any{
		"amount": "999.00",
		"reason": "B's own agreed figure",
	}, http.StatusOK, nil)

	before := workOrderOverride(t, srv.URL, tokenB, victim)
	if before == nil {
		t.Fatal("precondition failed: company B's override was not recorded")
	}

	// Company A tries to override company B's work order.
	doJSON(t, http.MethodPost, srv.URL+"/api/v1/work-orders/"+itoa(victim)+"/override-total",
		tokenA, map[string]any{"amount": "1.00", "reason": "not mine to set"},
		http.StatusNotFound, nil)

	after := workOrderOverride(t, srv.URL, tokenB, victim)
	if after == nil || *after != *before {
		t.Errorf("company B's override changed from %s to %s", deref(before), deref(after))
	}
}

// seedWorkOrder creates an asset and a work order in the company the token is
// scoped to, and returns the work order id.
func seedWorkOrder(t *testing.T, baseURL, token string, ts int64, tag string) int64 {
	t.Helper()

	var asset struct {
		ID int64 `json:"id"`
	}
	postJSON(t, baseURL+"/api/v1/assets", token, map[string]any{
		"name": "Unidad " + tag, "vin_sn": fmt.Sprintf("VIN-%s-%d", tag, ts),
	}, http.StatusCreated, &asset)

	var statuses struct {
		Data []struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	getJSON(t, baseURL+"/api/v1/work-order-statuses", token, http.StatusOK, &statuses)
	if len(statuses.Data) == 0 {
		t.Fatal("no seeded work order statuses; cannot create a work order")
	}

	var wo struct {
		ID int64 `json:"id"`
	}
	postJSON(t, baseURL+"/api/v1/work-orders", token, map[string]any{
		"asset_id": asset.ID, "status_id": statuses.Data[0].ID,
	}, http.StatusCreated, &wo)
	return wo.ID
}

// workOrderOverride reads back the stored total override, or nil when none is set.
func workOrderOverride(t *testing.T, baseURL, token string, id int64) *string {
	t.Helper()
	var wo struct {
		TotalOverride *string `json:"total_override"`
	}
	getJSON(t, baseURL+"/api/v1/work-orders/"+itoa(id), token, http.StatusOK, &wo)
	return wo.TotalOverride
}

func itoa(n int64) string { return fmt.Sprintf("%d", n) }

func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}
