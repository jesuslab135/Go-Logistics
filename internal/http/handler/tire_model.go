package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type TireModelStore struct{ q *gen.Queries }

func NewTireModelStore(q *gen.Queries) *TireModelStore { return &TireModelStore{q: q} }

func (s *TireModelStore) List(ctx context.Context, p paginate.Params) ([]dto.TireModelResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListTireModels(ctx, gen.ListTireModelsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountTireModels(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.TireModelResponse, len(rows))
	for i, r := range rows {
		out[i] = toTireModelResponse(r)
	}
	return out, total, nil
}

func (s *TireModelStore) Get(ctx context.Context, id int64) (dto.TireModelResponse, error) {
	r, err := s.q.GetTireModel(ctx, gen.GetTireModelParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.TireModelResponse{}, err
	}
	return toTireModelResponse(r), nil
}

func (s *TireModelStore) Create(ctx context.Context, in dto.CreateTireModelRequest) (dto.TireModelResponse, error) {
	r, err := s.q.CreateTireModel(ctx, gen.CreateTireModelParams{
		CompanyID:              middleware.CompanyFromContext(ctx),
		Brand:                  in.Brand,
		ModelName:              in.ModelName,
		Size:                   in.Size,
		FactoryTreadDepth32nds: in.FactoryTreadDepth32nds,
		MinimumTreadDepth32nds: in.MinimumTreadDepth32nds,
		LifeExpectancyMiles:    in.LifeExpectancyMiles,
		RecommendedPsi:         in.RecommendedPsi,
	})
	if err != nil {
		return dto.TireModelResponse{}, err
	}
	return toTireModelResponse(r), nil
}

func (s *TireModelStore) Update(ctx context.Context, id int64, in dto.UpdateTireModelRequest) (dto.TireModelResponse, error) {
	r, err := s.q.UpdateTireModel(ctx, gen.UpdateTireModelParams{
		ID:                     id,
		CompanyID:              middleware.CompanyFromContext(ctx),
		Brand:                  in.Brand,
		ModelName:              in.ModelName,
		Size:                   in.Size,
		FactoryTreadDepth32nds: in.FactoryTreadDepth32nds,
		MinimumTreadDepth32nds: in.MinimumTreadDepth32nds,
		LifeExpectancyMiles:    in.LifeExpectancyMiles,
		RecommendedPsi:         in.RecommendedPsi,
	})
	if err != nil {
		return dto.TireModelResponse{}, err
	}
	return toTireModelResponse(r), nil
}

func (s *TireModelStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteTireModel(ctx, gen.DeleteTireModelParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toTireModelResponse(r gen.TireModel) dto.TireModelResponse {
	return dto.TireModelResponse{
		ID:                     r.ID,
		CompanyID:              r.CompanyID,
		Brand:                  r.Brand,
		ModelName:              r.ModelName,
		Size:                   r.Size,
		FactoryTreadDepth32nds: r.FactoryTreadDepth32nds,
		MinimumTreadDepth32nds: r.MinimumTreadDepth32nds,
		LifeExpectancyMiles:    r.LifeExpectancyMiles,
		RecommendedPsi:         r.RecommendedPsi,
	}
}
