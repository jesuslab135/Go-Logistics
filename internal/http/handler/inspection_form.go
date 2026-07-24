package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type InspectionFormStore struct{ q *gen.Queries }

func NewInspectionFormStore(q *gen.Queries) *InspectionFormStore { return &InspectionFormStore{q: q} }

func (s *InspectionFormStore) List(ctx context.Context, p paginate.Params) ([]dto.InspectionFormResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListInspectionForms(ctx, gen.ListInspectionFormsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountInspectionForms(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.InspectionFormResponse, len(rows))
	for i, r := range rows {
		out[i] = toInspectionFormResponse(r)
	}
	return out, total, nil
}

func (s *InspectionFormStore) Get(ctx context.Context, id int64) (dto.InspectionFormResponse, error) {
	r, err := s.q.GetInspectionForm(ctx, gen.GetInspectionFormParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.InspectionFormResponse{}, err
	}
	return toInspectionFormResponse(r), nil
}

func (s *InspectionFormStore) Create(ctx context.Context, in dto.CreateInspectionFormRequest) (dto.InspectionFormResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateInspectionForm(ctx, gen.CreateInspectionFormParams{
		CompanyID:        middleware.CompanyFromContext(ctx),
		Title:            in.Title,
		Description:      in.Description,
		Version:          in.Version,
		RequireLivePhoto: in.RequireLivePhoto,
		AutoCreateIssues: in.AutoCreateIssues,
		Color:            in.Color,
		ArchivedAt:       in.ArchivedAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	})
	if err != nil {
		return dto.InspectionFormResponse{}, err
	}
	return toInspectionFormResponse(r), nil
}

func (s *InspectionFormStore) Update(ctx context.Context, id int64, in dto.UpdateInspectionFormRequest) (dto.InspectionFormResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateInspectionForm(ctx, gen.UpdateInspectionFormParams{
		ID:               id,
		CompanyID:        middleware.CompanyFromContext(ctx),
		Title:            in.Title,
		Description:      in.Description,
		Version:          in.Version,
		RequireLivePhoto: in.RequireLivePhoto,
		AutoCreateIssues: in.AutoCreateIssues,
		Color:            in.Color,
		ArchivedAt:       in.ArchivedAt,
		UpdatedAt:        now,
	})
	if err != nil {
		return dto.InspectionFormResponse{}, err
	}
	return toInspectionFormResponse(r), nil
}

func (s *InspectionFormStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteInspectionForm(ctx, gen.DeleteInspectionFormParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toInspectionFormResponse(r gen.InspectionForm) dto.InspectionFormResponse {
	return dto.InspectionFormResponse{
		ID:               r.ID,
		CompanyID:        r.CompanyID,
		Title:            r.Title,
		Description:      r.Description,
		Version:          r.Version,
		RequireLivePhoto: r.RequireLivePhoto,
		AutoCreateIssues: r.AutoCreateIssues,
		Color:            r.Color,
		ArchivedAt:       r.ArchivedAt,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}
