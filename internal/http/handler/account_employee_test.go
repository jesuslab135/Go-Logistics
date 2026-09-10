package handler

import (
	"errors"
	"strings"
	"testing"

	"fleet/internal/http/dto"
	"fleet/internal/platform/apierr"
)

// A grant naming the same company twice is ambiguous — the last one silently
// wins and the caller cannot tell which role was applied. Refuse it.
//
// The assertion is on the field-keyed Details map, not on Error(): that map is
// the actual contract render.go serializes to the client (see apierr.Error's
// json "details" field), while Error() carries no wrapped cause for a bare
// apierr.Validation and so is not part of what a caller observes.
func TestValidateGrantsRejectsDuplicateCompanies(t *testing.T) {
	in := dto.ReplaceAccountEmployeeCompaniesRequest{Grants: []dto.CompanyRoleGrant{
		{CompanyID: 1}, {CompanyID: 2}, {CompanyID: 1},
	}}
	err := validateGrants(in)

	var ae *apierr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("error = %v (%T), want *apierr.Error", err, err)
	}
	details, ok := ae.Details.(map[string]string)
	if !ok {
		t.Fatalf("details = %#v, want a map[string]string", ae.Details)
	}
	msg, ok := details["company_id"]
	if !ok {
		t.Fatalf("details = %#v, want a company_id entry", details)
	}
	if !strings.Contains(msg, "more than once") {
		t.Fatalf("company_id message = %q, want it to say the company is named more than once", msg)
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
