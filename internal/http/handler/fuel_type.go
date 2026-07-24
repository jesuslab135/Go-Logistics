package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type FuelTypeStore struct{ q *gen.Queries }

func NewFuelTypeStore(q *gen.Queries) *FuelTypeStore { return &FuelTypeStore{q: q} }

func (s *FuelTypeStore) List(ctx context.Context, p paginate.Params) ([]dto.FuelTypeResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListFuelTypes(ctx, gen.ListFuelTypesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountFuelTypes(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.FuelTypeResponse, len(rows))
	for i, r := range rows {
		out[i] = toFuelTypeResponse(r)
	}
	return out, total, nil
}

func (s *FuelTypeStore) Get(ctx context.Context, id int64) (dto.FuelTypeResponse, error) {
	r, err := s.q.GetFuelType(ctx, gen.GetFuelTypeParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.FuelTypeResponse{}, err
	}
	return toFuelTypeResponse(r), nil
}

func (s *FuelTypeStore) Create(ctx context.Context, in dto.CreateFuelTypeRequest) (dto.FuelTypeResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateFuelType(ctx, gen.CreateFuelTypeParams{
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		CreatedAt: now,
	})
	if err != nil {
		return dto.FuelTypeResponse{}, err
	}
	return toFuelTypeResponse(r), nil
}

func (s *FuelTypeStore) Update(ctx context.Context, id int64, in dto.UpdateFuelTypeRequest) (dto.FuelTypeResponse, error) {
	r, err := s.q.UpdateFuelType(ctx, gen.UpdateFuelTypeParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
	})
	if err != nil {
		return dto.FuelTypeResponse{}, err
	}
	return toFuelTypeResponse(r), nil
}

func (s *FuelTypeStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteFuelType(ctx, gen.DeleteFuelTypeParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toFuelTypeResponse(r gen.FuelType) dto.FuelTypeResponse {
	return dto.FuelTypeResponse{
		ID:        r.ID,
		CompanyID: r.CompanyID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt,
	}
}
