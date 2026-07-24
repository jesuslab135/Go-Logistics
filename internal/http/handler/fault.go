package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type FaultStore struct{ q *gen.Queries }

func NewFaultStore(q *gen.Queries) *FaultStore { return &FaultStore{q: q} }

func (s *FaultStore) List(ctx context.Context, p paginate.Params) ([]dto.FaultResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListFaults(ctx, gen.ListFaultsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountFaults(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.FaultResponse, len(rows))
	for i, r := range rows {
		out[i] = toFaultResponse(r)
	}
	return out, total, nil
}

func (s *FaultStore) Get(ctx context.Context, id int64) (dto.FaultResponse, error) {
	r, err := s.q.GetFault(ctx, gen.GetFaultParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.FaultResponse{}, err
	}
	return toFaultResponse(r), nil
}

func (s *FaultStore) Create(ctx context.Context, in dto.CreateFaultRequest) (dto.FaultResponse, error) {
	r, err := s.q.CreateFault(ctx, gen.CreateFaultParams{
		CompanyID:           middleware.CompanyFromContext(ctx),
		Family:              in.Family,
		Code:                in.Code,
		Name:                in.Name,
		Description:         in.Description,
		AppliesToAssetTypes: orEmptyStrings(in.AppliesToAssetTypes),
	})
	if err != nil {
		return dto.FaultResponse{}, err
	}
	return toFaultResponse(r), nil
}

func (s *FaultStore) Update(ctx context.Context, id int64, in dto.UpdateFaultRequest) (dto.FaultResponse, error) {
	r, err := s.q.UpdateFault(ctx, gen.UpdateFaultParams{
		ID:                  id,
		CompanyID:           middleware.CompanyFromContext(ctx),
		Family:              in.Family,
		Code:                in.Code,
		Name:                in.Name,
		Description:         in.Description,
		AppliesToAssetTypes: orEmptyStrings(in.AppliesToAssetTypes),
	})
	if err != nil {
		return dto.FaultResponse{}, err
	}
	return toFaultResponse(r), nil
}

func (s *FaultStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteFault(ctx, gen.DeleteFaultParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toFaultResponse(r gen.Fault) dto.FaultResponse {
	return dto.FaultResponse{
		ID:                  r.ID,
		CompanyID:           r.CompanyID,
		Family:              r.Family,
		Code:                r.Code,
		Name:                r.Name,
		Description:         r.Description,
		AppliesToAssetTypes: r.AppliesToAssetTypes,
	}
}
