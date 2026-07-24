package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

// ===================== VehicleMake =====================

type VehicleMakeStore struct{ q *gen.Queries }

func NewVehicleMakeStore(q *gen.Queries) *VehicleMakeStore { return &VehicleMakeStore{q: q} }

func (s *VehicleMakeStore) List(ctx context.Context, p paginate.Params) ([]dto.VehicleMakeResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListVehicleMakes(ctx, gen.ListVehicleMakesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountVehicleMakes(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.VehicleMakeResponse, len(rows))
	for i, r := range rows {
		out[i] = toVehicleMakeResponse(r)
	}
	return out, total, nil
}

func (s *VehicleMakeStore) Get(ctx context.Context, id int64) (dto.VehicleMakeResponse, error) {
	row, err := s.q.GetVehicleMake(ctx, gen.GetVehicleMakeParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.VehicleMakeResponse{}, err
	}
	return toVehicleMakeResponse(row), nil
}

func (s *VehicleMakeStore) Create(ctx context.Context, in dto.CreateVehicleMakeRequest) (dto.VehicleMakeResponse, error) {
	now := time.Now().UTC()
	row, err := s.q.CreateVehicleMake(ctx, gen.CreateVehicleMakeParams{
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return dto.VehicleMakeResponse{}, err
	}
	return toVehicleMakeResponse(row), nil
}

func (s *VehicleMakeStore) Update(ctx context.Context, id int64, in dto.UpdateVehicleMakeRequest) (dto.VehicleMakeResponse, error) {
	row, err := s.q.UpdateVehicleMake(ctx, gen.UpdateVehicleMakeParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return dto.VehicleMakeResponse{}, err
	}
	return toVehicleMakeResponse(row), nil
}

func (s *VehicleMakeStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteVehicleMake(ctx, gen.DeleteVehicleMakeParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toVehicleMakeResponse(r gen.VehicleMake) dto.VehicleMakeResponse {
	return dto.VehicleMakeResponse{ID: r.ID, CompanyID: r.CompanyID, Name: r.Name, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

// ===================== VehicleModel =====================

type VehicleModelStore struct{ q *gen.Queries }

func NewVehicleModelStore(q *gen.Queries) *VehicleModelStore { return &VehicleModelStore{q: q} }

func (s *VehicleModelStore) List(ctx context.Context, p paginate.Params) ([]dto.VehicleModelResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListVehicleModels(ctx, gen.ListVehicleModelsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountVehicleModels(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.VehicleModelResponse, len(rows))
	for i, r := range rows {
		out[i] = toVehicleModelResponse(r)
	}
	return out, total, nil
}

func (s *VehicleModelStore) Get(ctx context.Context, id int64) (dto.VehicleModelResponse, error) {
	row, err := s.q.GetVehicleModel(ctx, gen.GetVehicleModelParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.VehicleModelResponse{}, err
	}
	return toVehicleModelResponse(row), nil
}

func (s *VehicleModelStore) Create(ctx context.Context, in dto.CreateVehicleModelRequest) (dto.VehicleModelResponse, error) {
	now := time.Now().UTC()
	row, err := s.q.CreateVehicleModel(ctx, gen.CreateVehicleModelParams{
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		MakeID:    in.MakeID,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return dto.VehicleModelResponse{}, err
	}
	return toVehicleModelResponse(row), nil
}

func (s *VehicleModelStore) Update(ctx context.Context, id int64, in dto.UpdateVehicleModelRequest) (dto.VehicleModelResponse, error) {
	row, err := s.q.UpdateVehicleModel(ctx, gen.UpdateVehicleModelParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		MakeID:    in.MakeID,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return dto.VehicleModelResponse{}, err
	}
	return toVehicleModelResponse(row), nil
}

func (s *VehicleModelStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteVehicleModel(ctx, gen.DeleteVehicleModelParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toVehicleModelResponse(r gen.VehicleModel) dto.VehicleModelResponse {
	return dto.VehicleModelResponse{ID: r.ID, CompanyID: r.CompanyID, Name: r.Name, MakeID: r.MakeID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}
