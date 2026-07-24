package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type WarrantyStore struct{ q *gen.Queries }

func NewWarrantyStore(q *gen.Queries) *WarrantyStore { return &WarrantyStore{q: q} }

func (s *WarrantyStore) List(ctx context.Context, p paginate.Params) ([]dto.WarrantyResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListWarranties(ctx, gen.ListWarrantiesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountWarranties(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.WarrantyResponse, len(rows))
	for i, r := range rows {
		out[i] = toWarrantyResponse(r)
	}
	return out, total, nil
}

func (s *WarrantyStore) Get(ctx context.Context, id int64) (dto.WarrantyResponse, error) {
	r, err := s.q.GetWarranty(ctx, gen.GetWarrantyParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.WarrantyResponse{}, err
	}
	return toWarrantyResponse(r), nil
}

func (s *WarrantyStore) Create(ctx context.Context, in dto.CreateWarrantyRequest) (dto.WarrantyResponse, error) {
	r, err := s.q.CreateWarranty(ctx, gen.CreateWarrantyParams{
		CompanyID:  middleware.CompanyFromContext(ctx),
		ProviderID: in.ProviderID,
		AssetID:    in.AssetID,
		PartID:     in.PartID,
		StartDate:  in.StartDate,
		EndDate:    in.EndDate,
		Terms:      in.Terms,
		IsActive:   in.IsActive,
	})
	if err != nil {
		return dto.WarrantyResponse{}, err
	}
	return toWarrantyResponse(r), nil
}

func (s *WarrantyStore) Update(ctx context.Context, id int64, in dto.UpdateWarrantyRequest) (dto.WarrantyResponse, error) {
	r, err := s.q.UpdateWarranty(ctx, gen.UpdateWarrantyParams{
		ID:         id,
		CompanyID:  middleware.CompanyFromContext(ctx),
		ProviderID: in.ProviderID,
		AssetID:    in.AssetID,
		PartID:     in.PartID,
		StartDate:  in.StartDate,
		EndDate:    in.EndDate,
		Terms:      in.Terms,
		IsActive:   in.IsActive,
	})
	if err != nil {
		return dto.WarrantyResponse{}, err
	}
	return toWarrantyResponse(r), nil
}

func (s *WarrantyStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteWarranty(ctx, gen.DeleteWarrantyParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toWarrantyResponse(r gen.Warranty) dto.WarrantyResponse {
	return dto.WarrantyResponse{
		ID:         r.ID,
		CompanyID:  r.CompanyID,
		ProviderID: r.ProviderID,
		AssetID:    r.AssetID,
		PartID:     r.PartID,
		StartDate:  r.StartDate,
		EndDate:    r.EndDate,
		Terms:      r.Terms,
		IsActive:   r.IsActive,
	}
}
