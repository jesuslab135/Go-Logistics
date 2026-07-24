package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type MeasurementUnitStore struct{ q *gen.Queries }

func NewMeasurementUnitStore(q *gen.Queries) *MeasurementUnitStore {
	return &MeasurementUnitStore{q: q}
}

func (s *MeasurementUnitStore) List(ctx context.Context, p paginate.Params) ([]dto.MeasurementUnitResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListMeasurementUnits(ctx, gen.ListMeasurementUnitsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountMeasurementUnits(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.MeasurementUnitResponse, len(rows))
	for i, r := range rows {
		out[i] = toMeasurementUnitResponse(r)
	}
	return out, total, nil
}

func (s *MeasurementUnitStore) Get(ctx context.Context, id int64) (dto.MeasurementUnitResponse, error) {
	r, err := s.q.GetMeasurementUnit(ctx, gen.GetMeasurementUnitParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.MeasurementUnitResponse{}, err
	}
	return toMeasurementUnitResponse(r), nil
}

func (s *MeasurementUnitStore) Create(ctx context.Context, in dto.CreateMeasurementUnitRequest) (dto.MeasurementUnitResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateMeasurementUnit(ctx, gen.CreateMeasurementUnitParams{
		CompanyID:    middleware.CompanyFromContext(ctx),
		Name:         in.Name,
		Abbreviation: in.Abbreviation,
		CreatedAt:    now,
	})
	if err != nil {
		return dto.MeasurementUnitResponse{}, err
	}
	return toMeasurementUnitResponse(r), nil
}

func (s *MeasurementUnitStore) Update(ctx context.Context, id int64, in dto.UpdateMeasurementUnitRequest) (dto.MeasurementUnitResponse, error) {
	r, err := s.q.UpdateMeasurementUnit(ctx, gen.UpdateMeasurementUnitParams{
		ID:           id,
		CompanyID:    middleware.CompanyFromContext(ctx),
		Name:         in.Name,
		Abbreviation: in.Abbreviation,
	})
	if err != nil {
		return dto.MeasurementUnitResponse{}, err
	}
	return toMeasurementUnitResponse(r), nil
}

func (s *MeasurementUnitStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteMeasurementUnit(ctx, gen.DeleteMeasurementUnitParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toMeasurementUnitResponse(r gen.MeasurementUnit) dto.MeasurementUnitResponse {
	return dto.MeasurementUnitResponse{
		ID:           r.ID,
		CompanyID:    r.CompanyID,
		Name:         r.Name,
		Abbreviation: r.Abbreviation,
		CreatedAt:    r.CreatedAt,
	}
}
