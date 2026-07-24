package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type AxleTemplateStore struct{ q *gen.Queries }

func NewAxleTemplateStore(q *gen.Queries) *AxleTemplateStore { return &AxleTemplateStore{q: q} }

func (s *AxleTemplateStore) List(ctx context.Context, p paginate.Params) ([]dto.AxleTemplateResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListAxleTemplates(ctx, gen.ListAxleTemplatesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountAxleTemplates(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.AxleTemplateResponse, len(rows))
	for i, r := range rows {
		out[i] = toAxleTemplateResponse(r)
	}
	return out, total, nil
}

func (s *AxleTemplateStore) Get(ctx context.Context, id int64) (dto.AxleTemplateResponse, error) {
	r, err := s.q.GetAxleTemplate(ctx, gen.GetAxleTemplateParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.AxleTemplateResponse{}, err
	}
	return toAxleTemplateResponse(r), nil
}

func (s *AxleTemplateStore) Create(ctx context.Context, in dto.CreateAxleTemplateRequest) (dto.AxleTemplateResponse, error) {
	r, err := s.q.CreateAxleTemplate(ctx, gen.CreateAxleTemplateParams{
		CompanyID:      middleware.CompanyFromContext(ctx),
		Name:           in.Name,
		Description:    in.Description,
		TotalPositions: in.TotalPositions,
	})
	if err != nil {
		return dto.AxleTemplateResponse{}, err
	}
	return toAxleTemplateResponse(r), nil
}

func (s *AxleTemplateStore) Update(ctx context.Context, id int64, in dto.UpdateAxleTemplateRequest) (dto.AxleTemplateResponse, error) {
	r, err := s.q.UpdateAxleTemplate(ctx, gen.UpdateAxleTemplateParams{
		ID:             id,
		CompanyID:      middleware.CompanyFromContext(ctx),
		Name:           in.Name,
		Description:    in.Description,
		TotalPositions: in.TotalPositions,
	})
	if err != nil {
		return dto.AxleTemplateResponse{}, err
	}
	return toAxleTemplateResponse(r), nil
}

func (s *AxleTemplateStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteAxleTemplate(ctx, gen.DeleteAxleTemplateParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toAxleTemplateResponse(r gen.AxleTemplate) dto.AxleTemplateResponse {
	return dto.AxleTemplateResponse{
		ID:             r.ID,
		CompanyID:      r.CompanyID,
		Name:           r.Name,
		Description:    r.Description,
		TotalPositions: r.TotalPositions,
	}
}
