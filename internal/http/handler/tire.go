package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type TireStore struct{ q *gen.Queries }

func NewTireStore(q *gen.Queries) *TireStore { return &TireStore{q: q} }

func (s *TireStore) List(ctx context.Context, p paginate.Params) ([]dto.TireResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListTires(ctx, gen.ListTiresParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountTires(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.TireResponse, len(rows))
	for i, r := range rows {
		out[i] = toTireResponse(r)
	}
	return out, total, nil
}

func (s *TireStore) Get(ctx context.Context, id int64) (dto.TireResponse, error) {
	r, err := s.q.GetTire(ctx, gen.GetTireParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.TireResponse{}, err
	}
	return toTireResponse(r), nil
}

func (s *TireStore) Create(ctx context.Context, in dto.CreateTireRequest) (dto.TireResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateTire(ctx, gen.CreateTireParams{
		CompanyID:                middleware.CompanyFromContext(ctx),
		TireIdentificationNumber: in.TireIdentificationNumber,
		TireModelID:              in.TireModelID,
		Status:                   in.Status,
		CurrentTreadDepth32nds:   in.CurrentTreadDepth32nds,
		CurrentPsi:               in.CurrentPsi,
		TotalMiles:               in.TotalMiles,
		CurrentVehicleID:         in.CurrentVehicleID,
		CurrentPositionCode:      in.CurrentPositionCode,
		PurchaseDate:             in.PurchaseDate,
		PurchaseCost:             in.PurchaseCost,
		VendorID:                 in.VendorID,
		CreatedAt:                now,
	})
	if err != nil {
		return dto.TireResponse{}, err
	}
	return toTireResponse(r), nil
}

func (s *TireStore) Update(ctx context.Context, id int64, in dto.UpdateTireRequest) (dto.TireResponse, error) {
	r, err := s.q.UpdateTire(ctx, gen.UpdateTireParams{
		ID:                       id,
		CompanyID:                middleware.CompanyFromContext(ctx),
		TireIdentificationNumber: in.TireIdentificationNumber,
		TireModelID:              in.TireModelID,
		Status:                   in.Status,
		CurrentTreadDepth32nds:   in.CurrentTreadDepth32nds,
		CurrentPsi:               in.CurrentPsi,
		TotalMiles:               in.TotalMiles,
		CurrentVehicleID:         in.CurrentVehicleID,
		CurrentPositionCode:      in.CurrentPositionCode,
		PurchaseDate:             in.PurchaseDate,
		PurchaseCost:             in.PurchaseCost,
		VendorID:                 in.VendorID,
	})
	if err != nil {
		return dto.TireResponse{}, err
	}
	return toTireResponse(r), nil
}

func (s *TireStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteTire(ctx, gen.DeleteTireParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toTireResponse(r gen.Tire) dto.TireResponse {
	return dto.TireResponse{
		ID:                       r.ID,
		CompanyID:                r.CompanyID,
		TireIdentificationNumber: r.TireIdentificationNumber,
		TireModelID:              r.TireModelID,
		Status:                   r.Status,
		CurrentTreadDepth32nds:   r.CurrentTreadDepth32nds,
		CurrentPsi:               r.CurrentPsi,
		TotalMiles:               r.TotalMiles,
		CurrentVehicleID:         r.CurrentVehicleID,
		CurrentPositionCode:      r.CurrentPositionCode,
		PurchaseDate:             r.PurchaseDate,
		PurchaseCost:             r.PurchaseCost,
		VendorID:                 r.VendorID,
		CreatedAt:                r.CreatedAt,
	}
}
