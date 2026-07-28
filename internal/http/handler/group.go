package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type GroupStore struct{ q *gen.Queries }

func NewGroupStore(q *gen.Queries) *GroupStore { return &GroupStore{q: q} }

func (s *GroupStore) List(ctx context.Context, p paginate.Params) ([]dto.GroupResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListGroups(ctx, gen.ListGroupsParams{CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountGroups(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.GroupResponse, len(rows))
	for i, r := range rows {
		out[i] = toGroupResponse(r)
	}
	return out, total, nil
}

func (s *GroupStore) Get(ctx context.Context, id int64) (dto.GroupResponse, error) {
	r, err := s.q.GetGroup(ctx, gen.GetGroupParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.GroupResponse{}, err
	}
	return toGroupResponse(r), nil
}

func (s *GroupStore) Create(ctx context.Context, in dto.CreateGroupRequest) (dto.GroupResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateGroup(ctx, gen.CreateGroupParams{
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		ParentID:  in.ParentID,
		IsDefault: in.IsDefault,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return dto.GroupResponse{}, err
	}
	return toGroupResponse(r), nil
}

func (s *GroupStore) Update(ctx context.Context, id int64, in dto.UpdateGroupRequest) (dto.GroupResponse, error) {
	r, err := s.q.UpdateGroup(ctx, gen.UpdateGroupParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		ParentID:  in.ParentID,
		IsDefault: in.IsDefault,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return dto.GroupResponse{}, err
	}
	return toGroupResponse(r), nil
}

func (s *GroupStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteGroup(ctx, gen.DeleteGroupParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toGroupResponse(r gen.Group) dto.GroupResponse {
	return dto.GroupResponse{
		ID:        r.ID,
		CompanyID: r.CompanyID,
		Name:      r.Name,
		ParentID:  r.ParentID,
		Ancestry:  r.Ancestry,
		IsDefault: r.IsDefault,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
