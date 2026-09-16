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

// adminCompany is the subset of dto.AdminCompanyResponse these tests assert on.
type adminCompany struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	AccountID     int64  `json:"account_id"`
	AccountName   string `json:"account_name"`
	EmployeeCount int64  `json:"employee_count"`
}

func TestAdminListsCompaniesAcrossAccounts(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	srv := httptest.NewServer(newIntegrationRouter(pool))
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	accountA, _, _, ownerA := provisionAccountOwner(t, srv.URL, platformToken)
	accountB, _, _, ownerB := provisionAccountOwner(t, srv.URL, platformToken)
	ts := time.Now().UnixNano()
	companyA := createCompany(t, srv.URL, ownerA, "Alpha Fleet", fmt.Sprintf("ALPHA-%d", ts))
	companyB := createCompany(t, srv.URL, ownerB, "Beta Fleet", fmt.Sprintf("BETA-%d", ts))

	// Unfiltered: both tenants, which no tenant-scoped route can show.
	var all struct {
		Data  []adminCompany `json:"data"`
		Total int64          `json:"total"`
	}
	getJSON(t, srv.URL+"/api/v1/admin/companies", platformToken, http.StatusOK, &all)
	if all.Total != 2 || len(all.Data) != 2 {
		t.Fatalf("unfiltered list = %+v, want both companies", all)
	}

	// Filtered to one client.
	var onlyA struct {
		Data  []adminCompany `json:"data"`
		Total int64          `json:"total"`
	}
	getJSON(t, fmt.Sprintf("%s/api/v1/admin/companies?account_id=%d", srv.URL, accountA), platformToken, http.StatusOK, &onlyA)
	if onlyA.Total != 1 || len(onlyA.Data) != 1 {
		t.Fatalf("account_id filter = %+v, want exactly company A", onlyA)
	}
	a := onlyA.Data[0]
	if a.ID != companyA || a.AccountID != accountA || a.AccountName != "Integration Fleet" || a.EmployeeCount != 1 {
		t.Fatalf("company A row = %+v, want id %d, account %d, account_name Integration Fleet, 1 employee", a, companyA, accountA)
	}

	// Detail.
	var b adminCompany
	getJSON(t, fmt.Sprintf("%s/api/v1/admin/companies/%d", srv.URL, companyB), platformToken, http.StatusOK, &b)
	if b.ID != companyB || b.AccountID != accountB || b.Name != "Beta Fleet" {
		t.Fatalf("company B detail = %+v, want id %d in account %d", b, companyB, accountB)
	}

	getJSON(t, srv.URL+"/api/v1/admin/companies/999999", platformToken, http.StatusNotFound, nil)
	getJSON(t, srv.URL+"/api/v1/admin/companies?account_id=abc", platformToken, http.StatusBadRequest, nil)

	// A client owner is not platform staff.
	getJSON(t, srv.URL+"/api/v1/admin/companies", ownerA, http.StatusForbidden, nil)
}
