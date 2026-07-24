package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
)

type VehicleAxleConfigStore struct{ q *gen.Queries }

func NewVehicleAxleConfigStore(q *gen.Queries) *VehicleAxleConfigStore {
	return &VehicleAxleConfigStore{q: q}
}

func (s *VehicleAxleConfigStore) Get(ctx context.Context, parentID int64) (dto.VehicleAxleConfigResponse, error) {
	r, err := s.q.GetVehicleAxleConfig(ctx, gen.GetVehicleAxleConfigParams{ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.VehicleAxleConfigResponse{}, err
	}
	return toVehicleAxleConfigResponse(r), nil
}

func (s *VehicleAxleConfigStore) Upsert(ctx context.Context, parentID int64, in dto.UpsertVehicleAxleConfigRequest) (dto.VehicleAxleConfigResponse, error) {
	r, err := s.q.UpsertVehicleAxleConfig(ctx, gen.UpsertVehicleAxleConfigParams{
		ParentID:    parentID,
		CompanyID:   middleware.CompanyFromContext(ctx),
		TemplateID:  in.TemplateID,
		DisplayName: in.DisplayName,
	})
	if err != nil {
		return dto.VehicleAxleConfigResponse{}, err
	}
	return toVehicleAxleConfigResponse(r), nil
}

func (s *VehicleAxleConfigStore) Delete(ctx context.Context, parentID int64) error {
	return s.q.DeleteVehicleAxleConfig(ctx, gen.DeleteVehicleAxleConfigParams{ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toVehicleAxleConfigResponse(r gen.VehicleAxleConfig) dto.VehicleAxleConfigResponse {
	return dto.VehicleAxleConfigResponse{
		VehicleID:   r.VehicleID,
		TemplateID:  r.TemplateID,
		DisplayName: r.DisplayName,
	}
}
