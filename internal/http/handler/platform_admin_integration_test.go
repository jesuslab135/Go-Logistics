//go:build integration

package handler

import (
	"context"
	"encoding/json"
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

// adminAccount is the subset of dto.AccountResponse these tests assert on.
type adminAccount struct {
	ID              int64  `json:"id"`
	OwnerEmployeeID *int64 `json:"owner_employee_id"`
	OwnerEmail      string `json:"owner_email"`
	CompanyCount    int64  `json:"company_count"`
	EmployeeCount   int64  `json:"employee_count"`
}

func TestAdminAccountDetailAndSetOwnerCarryCounts(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	srv := httptest.NewServer(newIntegrationRouter(pool))
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	accountID, ownerID, ownerEmail, ownerToken := provisionAccountOwner(t, srv.URL, platformToken)
	createCompany(t, srv.URL, ownerToken, "Gamma Fleet", fmt.Sprintf("GAMMA-%d", time.Now().UnixNano()))

	var detail adminAccount
	getJSON(t, fmt.Sprintf("%s/api/v1/admin/accounts/%d", srv.URL, accountID), platformToken, http.StatusOK, &detail)
	if detail.ID != accountID || detail.OwnerEmail != ownerEmail || detail.CompanyCount != 1 || detail.EmployeeCount != 1 {
		t.Fatalf("account detail = %+v, want id %d, owner %s, 1 company, 1 employee", detail, accountID, ownerEmail)
	}
	getJSON(t, srv.URL+"/api/v1/admin/accounts/999999", platformToken, http.StatusNotFound, nil)
	getJSON(t, fmt.Sprintf("%s/api/v1/admin/accounts/%d", srv.URL, accountID), ownerToken, http.StatusForbidden, nil)

	// Re-affirming the current owner is a valid transfer; the response must be
	// the same complete account the detail route returns, not a partial one.
	var afterSet adminAccount
	postJSON(t, fmt.Sprintf("%s/api/v1/admin/accounts/%d/set-owner", srv.URL, accountID), platformToken,
		map[string]any{"employee_id": ownerID}, http.StatusOK, &afterSet)
	if afterSet.OwnerEmail != ownerEmail || afterSet.CompanyCount != 1 || afterSet.EmployeeCount != 1 {
		t.Fatalf("set-owner response = %+v, want owner %s with counts 1/1", afterSet, ownerEmail)
	}
}

// meEmployeeID reads the signed-in employee's id off /me/permissions.
func meEmployeeID(t *testing.T, baseURL, token string) int64 {
	t.Helper()
	var me struct {
		Employee struct {
			ID int64 `json:"id"`
		} `json:"employee"`
	}
	getJSON(t, baseURL+"/api/v1/me/permissions", token, http.StatusOK, &me)
	if me.Employee.ID == 0 {
		t.Fatal("/me/permissions returned no employee id")
	}
	return me.Employee.ID
}

type adminEmployee struct {
	ID              int64  `json:"id"`
	AccountID       *int64 `json:"account_id"`
	IsPlatformAdmin bool   `json:"is_platform_admin"`
}

func TestAdminEmployeesSayWhichClientTheyBelongTo(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	srv := httptest.NewServer(newIntegrationRouter(pool))
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	platformID := meEmployeeID(t, srv.URL, platformToken)
	accountA, ownerA, _, _ := provisionAccountOwner(t, srv.URL, platformToken)
	_, ownerB, _, _ := provisionAccountOwner(t, srv.URL, platformToken)

	var all struct {
		Data  []adminEmployee `json:"data"`
		Total int64           `json:"total"`
	}
	getJSON(t, srv.URL+"/api/v1/admin/employees", platformToken, http.StatusOK, &all)
	if all.Total != 3 {
		t.Fatalf("unfiltered total = %d, want 3 (platform admin + two owners)", all.Total)
	}
	byID := map[int64]adminEmployee{}
	for _, e := range all.Data {
		byID[e.ID] = e
	}
	if p := byID[platformID]; p.AccountID != nil || !p.IsPlatformAdmin {
		t.Fatalf("platform admin row = %+v, want no account and the flag", p)
	}
	if a := byID[ownerA]; a.AccountID == nil || *a.AccountID != accountA || a.IsPlatformAdmin {
		t.Fatalf("owner A row = %+v, want account %d", a, accountA)
	}

	var onlyA struct {
		Data  []adminEmployee `json:"data"`
		Total int64           `json:"total"`
	}
	getJSON(t, fmt.Sprintf("%s/api/v1/admin/employees?account_id=%d", srv.URL, accountA), platformToken, http.StatusOK, &onlyA)
	if onlyA.Total != 1 || len(onlyA.Data) != 1 || onlyA.Data[0].ID != ownerA {
		t.Fatalf("account_id filter = %+v, want only owner A (%d), not owner B (%d)", onlyA, ownerA, ownerB)
	}

	getJSON(t, srv.URL+"/api/v1/admin/employees?account_id=x", platformToken, http.StatusBadRequest, nil)
}

func TestMembershipAcrossClientsIsA422NamingCompanyIDs(t *testing.T) {
	ctx := context.Background()
	pool := setupThrowawayDB(t, ctx)
	srv := httptest.NewServer(newIntegrationRouter(pool))
	defer srv.Close()

	platformToken := seedPlatformAdmin(t, ctx, pool, srv.URL)
	platformID := meEmployeeID(t, srv.URL, platformToken)
	_, ownerA, _, tokenA := provisionAccountOwner(t, srv.URL, platformToken)
	_, _, _, tokenB := provisionAccountOwner(t, srv.URL, platformToken)
	ts := time.Now().UnixNano()
	companyA := createCompany(t, srv.URL, tokenA, "Delta Fleet", fmt.Sprintf("DELTA-%d", ts))
	companyB := createCompany(t, srv.URL, tokenB, "Epsilon Fleet", fmt.Sprintf("EPS-%d", ts))

	namesCompanyIDs := func(body []byte) {
		t.Helper()
		var e struct {
			Error struct {
				Code    string            `json:"code"`
				Details map[string]string `json:"details"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &e); err != nil {
			t.Fatalf("decode error body: %v (%s)", err, body)
		}
		if e.Error.Code != "validation_failed" || e.Error.Details["company_ids"] == "" {
			t.Fatalf("error = %s, want validation_failed naming company_ids", body)
		}
	}

	// Owner A into client B's company.
	body := putJSON(t, fmt.Sprintf("%s/api/v1/admin/employees/%d/companies", srv.URL, ownerA), platformToken,
		map[string]any{"company_ids": []int64{companyA, companyB}, "default_company_id": companyA},
		http.StatusUnprocessableEntity, nil)
	namesCompanyIDs(body)

	// Nothing changed.
	var memberships struct {
		CompanyIDs []int64 `json:"company_ids"`
	}
	getJSON(t, fmt.Sprintf("%s/api/v1/admin/employees/%d/companies", srv.URL, ownerA), platformToken, http.StatusOK, &memberships)
	if len(memberships.CompanyIDs) != 1 || memberships.CompanyIDs[0] != companyA {
		t.Fatalf("owner A memberships = %v, want unchanged [%d]", memberships.CompanyIDs, companyA)
	}

	// Platform staff into any company.
	body = putJSON(t, fmt.Sprintf("%s/api/v1/admin/employees/%d/companies", srv.URL, platformID), platformToken,
		map[string]any{"company_ids": []int64{companyA}, "default_company_id": companyA},
		http.StatusUnprocessableEntity, nil)
	namesCompanyIDs(body)

	// The valid case still works.
	putJSON(t, fmt.Sprintf("%s/api/v1/admin/employees/%d/companies", srv.URL, ownerA), platformToken,
		map[string]any{"company_ids": []int64{companyA}, "default_company_id": companyA},
		http.StatusOK, nil)
}
