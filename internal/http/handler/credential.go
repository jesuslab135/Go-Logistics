package handler

import (
	"context"
	"slices"

	"fleet/internal/auth"
	"fleet/internal/db/gen"
)

var _ CredentialVerifier = (*EmployeeCredentialVerifier)(nil)

// EmployeeCredentialVerifier authenticates against the employee table's bcrypt
// password_hash, resolving the tenant from default_company_id and admin status
// from the linked role.
type EmployeeCredentialVerifier struct {
	q *gen.Queries
}

func NewEmployeeCredentialVerifier(q *gen.Queries) *EmployeeCredentialVerifier {
	return &EmployeeCredentialVerifier{q: q}
}

func (v *EmployeeCredentialVerifier) Verify(ctx context.Context, email, password string) (Identity, error) {
	row, err := v.q.GetEmployeeAuthByEmail(ctx, email)
	if err != nil {
		// Unknown email (ErrNoRows) or lookup failure — do not distinguish, to
		// avoid leaking which emails exist.
		return Identity{}, ErrInvalidCredentials
	}
	if !auth.CheckPassword(row.PasswordHash, password) {
		return Identity{}, ErrInvalidCredentials
	}

	// default_company_id is a plain writable field, so it is a hint, not an
	// authority: the session is scoped to a company only after employee_companies
	// confirms the membership — the same check /auth/switch-company already makes.
	memberships, err := v.q.ListEmployeeCompanyIDs(ctx, row.ID)
	if err != nil {
		return Identity{}, ErrInvalidCredentials
	}
	companyID, ok := resolveLoginCompany(row.DefaultCompanyID, memberships)
	if !ok {
		return Identity{}, ErrNoCompanyMembership
	}

	return Identity{EmployeeID: row.ID, CompanyID: companyID, IsAdmin: row.IsAdmin}, nil
}

// resolveLoginCompany picks the company a session is scoped to. The employee's
// default_company_id wins when it names a real membership; otherwise the lowest
// membership id does, so the choice is deterministic across logins. An employee
// with no membership resolves to nothing and must not be issued a token.
func resolveLoginCompany(defaultCompanyID *int64, memberships []int64) (int64, bool) {
	if len(memberships) == 0 {
		return 0, false
	}
	if defaultCompanyID != nil && slices.Contains(memberships, *defaultCompanyID) {
		return *defaultCompanyID, true
	}
	return slices.Min(memberships), true
}
