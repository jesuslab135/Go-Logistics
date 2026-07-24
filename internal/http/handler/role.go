package handler

import (
	"context"
	"encoding/json"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type RoleStore struct{ q *gen.Queries }

func NewRoleStore(q *gen.Queries) *RoleStore { return &RoleStore{q: q} }

func (s *RoleStore) List(ctx context.Context, p paginate.Params) ([]dto.RoleResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListRoles(ctx, gen.ListRolesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountRoles(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.RoleResponse, len(rows))
	for i, r := range rows {
		out[i] = toRoleResponse(r)
	}
	return out, total, nil
}

func (s *RoleStore) Get(ctx context.Context, id int64) (dto.RoleResponse, error) {
	r, err := s.q.GetRole(ctx, gen.GetRoleParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.RoleResponse{}, err
	}
	return toRoleResponse(r), nil
}

func (s *RoleStore) Create(ctx context.Context, in dto.CreateRoleRequest) (dto.RoleResponse, error) {
	r, err := s.q.CreateRole(ctx, gen.CreateRoleParams{
		CompanyID:   middleware.CompanyFromContext(ctx),
		Name:        in.Name,
		IsAdmin:     in.IsAdmin,
		Permissions: jsonbOrDefault(in.Permissions, "{}"),
	})
	if err != nil {
		return dto.RoleResponse{}, err
	}
	return toRoleResponse(r), nil
}

func (s *RoleStore) Update(ctx context.Context, id int64, in dto.UpdateRoleRequest) (dto.RoleResponse, error) {
	r, err := s.q.UpdateRole(ctx, gen.UpdateRoleParams{
		ID:          id,
		CompanyID:   middleware.CompanyFromContext(ctx),
		Name:        in.Name,
		IsAdmin:     in.IsAdmin,
		Permissions: jsonbOrDefault(in.Permissions, "{}"),
	})
	if err != nil {
		return dto.RoleResponse{}, err
	}
	return toRoleResponse(r), nil
}

func (s *RoleStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteRole(ctx, gen.DeleteRoleParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toRoleResponse(r gen.Role) dto.RoleResponse {
	return dto.RoleResponse{
		ID:          r.ID,
		CompanyID:   r.CompanyID,
		Name:        r.Name,
		IsAdmin:     r.IsAdmin,
		Permissions: json.RawMessage(r.Permissions),
	}
}
