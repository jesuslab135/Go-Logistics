package handler

import (
	"context"

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

	var companyID int64
	if row.DefaultCompanyID != nil {
		companyID = *row.DefaultCompanyID
	}
	return Identity{EmployeeID: row.ID, CompanyID: companyID, IsAdmin: row.IsAdmin}, nil
}
