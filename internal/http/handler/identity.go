package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/middleware"
)

// IdentityLoader resolves the authorization subject for each request, standing
// in for Django's request.user.employee lookup.
type IdentityLoader struct {
	q *gen.Queries
}

func NewIdentityLoader(q *gen.Queries) *IdentityLoader {
	return &IdentityLoader{q: q}
}

func (l *IdentityLoader) LoadIdentity(ctx context.Context, employeeID int64, companyID *int64) (middleware.Identity, error) {
	row, err := l.q.GetEmployeeIdentity(ctx, gen.GetEmployeeIdentityParams{
		ID:        employeeID,
		CompanyID: companyID,
	})
	if err != nil {
		return middleware.Identity{}, err
	}

	// A company-less session (no tenant on the token) reports CompanyID 0, the
	// same "absent" convention CompanyFromContext already uses; IsMember is
	// false for it regardless, since the identity query's membership EXISTS
	// resolves to false when companyID is nil.
	var company int64
	if companyID != nil {
		company = *companyID
	}

	return middleware.Identity{
		EmployeeID:     row.ID,
		CompanyID:      company,
		AccountID:      row.AccountID,
		IsActive:       row.IsActive,
		IsAccountOwner: row.IsAccountOwner,
		IsMember:       row.IsMember,
		HasRole:        row.RoleID != nil,
		// Administrator of THIS company, which needs membership in it. The query's
		// is_admin is also true for an account owner with no company yet (from
		// ownership alone), and reporting that would make /me/permissions list
		// every module while every module route refuses a non-member.
		IsAdmin:         row.IsAdmin && row.IsMember,
		IsPlatformAdmin: row.IsPlatformAdmin,
		RoleID:          row.RoleID,
		Permissions:     middleware.DecodePermissions(row.Permissions),
	}, nil
}
