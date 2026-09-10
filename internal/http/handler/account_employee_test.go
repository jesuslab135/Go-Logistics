package handler

import (
	"strings"
	"testing"

	"fleet/internal/http/dto"
)

// A grant naming the same company twice is ambiguous — the last one silently
// wins and the caller cannot tell which role was applied. Refuse it.
func TestValidateGrantsRejectsDuplicateCompanies(t *testing.T) {
	in := dto.ReplaceAccountEmployeeCompaniesRequest{Grants: []dto.CompanyRoleGrant{
		{CompanyID: 1}, {CompanyID: 2}, {CompanyID: 1},
	}}
	err := validateGrants(in)
	if err == nil || !strings.Contains(err.Error(), "company") {
		t.Fatalf("error = %v, want it to name the duplicate company", err)
	}
}

func TestValidateGrantsRejectsNonPositiveCompany(t *testing.T) {
	in := dto.ReplaceAccountEmployeeCompaniesRequest{Grants: []dto.CompanyRoleGrant{{CompanyID: 0}}}
	if err := validateGrants(in); err == nil {
		t.Fatal("a company id of 0 must be refused")
	}
}

// An empty grant list is legitimate: it revokes every membership. It must not
// be confused with a malformed request.
func TestValidateGrantsAllowsEmpty(t *testing.T) {
	if err := validateGrants(dto.ReplaceAccountEmployeeCompaniesRequest{}); err != nil {
		t.Fatalf("an empty grant list revokes all memberships and is valid: %v", err)
	}
}
