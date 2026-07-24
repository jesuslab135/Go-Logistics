package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type PartCategoryStore struct{ q *gen.Queries }

func NewPartCategoryStore(q *gen.Queries) *PartCategoryStore { return &PartCategoryStore{q: q} }

func (s *PartCategoryStore) List(ctx context.Context, p paginate.Params) ([]dto.PartCategoryResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListPartCategories(ctx, gen.ListPartCategoriesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountPartCategories(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PartCategoryResponse, len(rows))
	for i, r := range rows {
		out[i] = toPartCategoryResponse(r)
	}
	return out, total, nil
}

func (s *PartCategoryStore) Get(ctx context.Context, id int64) (dto.PartCategoryResponse, error) {
	r, err := s.q.GetPartCategory(ctx, gen.GetPartCategoryParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.PartCategoryResponse{}, err
	}
	return toPartCategoryResponse(r), nil
}

func (s *PartCategoryStore) Create(ctx context.Context, in dto.CreatePartCategoryRequest) (dto.PartCategoryResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreatePartCategory(ctx, gen.CreatePartCategoryParams{
		CompanyID:   middleware.CompanyFromContext(ctx),
		Name:        in.Name,
		Description: in.Description,
		CreatedAt:   now,
	})
	if err != nil {
		return dto.PartCategoryResponse{}, err
	}
	return toPartCategoryResponse(r), nil
}

func (s *PartCategoryStore) Update(ctx context.Context, id int64, in dto.UpdatePartCategoryRequest) (dto.PartCategoryResponse, error) {
	r, err := s.q.UpdatePartCategory(ctx, gen.UpdatePartCategoryParams{
		ID:          id,
		CompanyID:   middleware.CompanyFromContext(ctx),
		Name:        in.Name,
		Description: in.Description,
	})
	if err != nil {
		return dto.PartCategoryResponse{}, err
	}
	return toPartCategoryResponse(r), nil
}

func (s *PartCategoryStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeletePartCategory(ctx, gen.DeletePartCategoryParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toPartCategoryResponse(r gen.PartCategory) dto.PartCategoryResponse {
	return dto.PartCategoryResponse{
		ID:          r.ID,
		CompanyID:   r.CompanyID,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
	}
}
