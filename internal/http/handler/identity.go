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

func (l *IdentityLoader) LoadIdentity(ctx context.Context, employeeID, companyID int64) (middleware.Identity, error) {
	row, err := l.q.GetEmployeeIdentity(ctx, gen.GetEmployeeIdentityParams{
		ID:        employeeID,
		CompanyID: companyID,
	})
	if err != nil {
		return middleware.Identity{}, err
	}

	return middleware.Identity{
		EmployeeID:     row.ID,
		CompanyID:      companyID,
		IsActive:       row.IsActive,
		IsAccountOwner: row.IsAccountOwner,
		IsMember:       row.IsMember,
		HasRole:        row.RoleID != nil,
		IsAdmin:        row.IsAdmin,
		Permissions:    middleware.DecodePermissions(row.Permissions),
	}, nil
}
