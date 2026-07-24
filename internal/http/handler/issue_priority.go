package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type IssuePriorityStore struct{ q *gen.Queries }

func NewIssuePriorityStore(q *gen.Queries) *IssuePriorityStore { return &IssuePriorityStore{q: q} }

func (s *IssuePriorityStore) List(ctx context.Context, p paginate.Params) ([]dto.IssuePriorityResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListIssuePriorities(ctx, gen.ListIssuePrioritiesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountIssuePriorities(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.IssuePriorityResponse, len(rows))
	for i, r := range rows {
		out[i] = toIssuePriorityResponse(r)
	}
	return out, total, nil
}

func (s *IssuePriorityStore) Get(ctx context.Context, id int64) (dto.IssuePriorityResponse, error) {
	r, err := s.q.GetIssuePriority(ctx, gen.GetIssuePriorityParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.IssuePriorityResponse{}, err
	}
	return toIssuePriorityResponse(r), nil
}

func (s *IssuePriorityStore) Create(ctx context.Context, in dto.CreateIssuePriorityRequest) (dto.IssuePriorityResponse, error) {
	r, err := s.q.CreateIssuePriority(ctx, gen.CreateIssuePriorityParams{
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		Color:     in.Color,
		Position:  in.Position,
	})
	if err != nil {
		return dto.IssuePriorityResponse{}, err
	}
	return toIssuePriorityResponse(r), nil
}

func (s *IssuePriorityStore) Update(ctx context.Context, id int64, in dto.UpdateIssuePriorityRequest) (dto.IssuePriorityResponse, error) {
	r, err := s.q.UpdateIssuePriority(ctx, gen.UpdateIssuePriorityParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		Color:     in.Color,
		Position:  in.Position,
	})
	if err != nil {
		return dto.IssuePriorityResponse{}, err
	}
	return toIssuePriorityResponse(r), nil
}

func (s *IssuePriorityStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteIssuePriority(ctx, gen.DeleteIssuePriorityParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toIssuePriorityResponse(r gen.IssuePriority) dto.IssuePriorityResponse {
	return dto.IssuePriorityResponse{
		ID:        r.ID,
		CompanyID: r.CompanyID,
		Name:      r.Name,
		Color:     r.Color,
		Position:  r.Position,
	}
}
