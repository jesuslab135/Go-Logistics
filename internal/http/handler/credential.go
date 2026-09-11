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
	companyID := resolveLoginCompany(row.DefaultCompanyID, memberships)

	if !mayLogIn(companyID, row.AccountID) {
		return Identity{}, ErrNoCompanyMembership
	}

	// is_admin is a property of the role held in THIS company, so it cannot be
	// resolved until the company is. It needs membership in that company, the
	// same rule IdentityLoader applies at request time: a company-less session
	// is admin of nothing, owner or not, and the token claim must not say
	// otherwise.
	ident, err := v.q.GetEmployeeIdentity(ctx, gen.GetEmployeeIdentityParams{
		ID:        row.ID,
		CompanyID: companyID,
	})
	if err != nil {
		return Identity{}, ErrInvalidCredentials
	}

	return Identity{
		EmployeeID: row.ID,
		CompanyID:  companyID,
		AccountID:  row.AccountID,
		IsAdmin:    ident.IsAdmin && ident.IsMember,
	}, nil
}

// resolveLoginCompany picks the company a session is scoped to. The employee's
// default_company_id wins when it names a real membership; otherwise the lowest
// membership id does, so the choice is deterministic across logins.
//
// An employee with no membership resolves to nil, which is a company-less
// session rather than a refusal: an account owner provisioned by a platform
// admin has no company until they create one, and refusing them a token is the
// deadlock this change exists to break. Whether they may log in at all is
// decided by Verify, on whether they belong to an account.
func resolveLoginCompany(defaultCompanyID *int64, memberships []int64) *int64 {
	if len(memberships) == 0 {
		return nil
	}
	if defaultCompanyID != nil && slices.Contains(memberships, *defaultCompanyID) {
		return defaultCompanyID
	}
	lowest := slices.Min(memberships)
	return &lowest
}

// mayLogIn reports whether an employee may receive a token at all.
//
// A company is not required: an account owner provisioned by a platform admin
// has none until they create one, and refusing them a token is the deadlock
// this change exists to break. Belonging to NOTHING is still refused — a nil
// company and a nil account together mean there is nowhere for the session to
// be.
func mayLogIn(companyID *int64, accountID *int64) bool {
	return companyID != nil || accountID != nil
}
