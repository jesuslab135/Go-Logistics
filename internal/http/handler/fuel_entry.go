package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type FuelEntryStore struct{ q *gen.Queries }

func NewFuelEntryStore(q *gen.Queries) *FuelEntryStore { return &FuelEntryStore{q: q} }

func (s *FuelEntryStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.FuelEntryResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListFuelEntries(ctx, gen.ListFuelEntriesParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountFuelEntries(ctx, gen.CountFuelEntriesParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.FuelEntryResponse, len(rows))
	for i, r := range rows {
		out[i] = toFuelEntryResponse(r)
	}
	return out, total, nil
}

func (s *FuelEntryStore) Get(ctx context.Context, parentID, id int64) (dto.FuelEntryResponse, error) {
	r, err := s.q.GetFuelEntry(ctx, gen.GetFuelEntryParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.FuelEntryResponse{}, err
	}
	return toFuelEntryResponse(r), nil
}

func (s *FuelEntryStore) Create(ctx context.Context, parentID int64, in dto.CreateFuelEntryRequest) (dto.FuelEntryResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateFuelEntry(ctx, gen.CreateFuelEntryParams{
		ParentID:       parentID,
		CompanyID:      middleware.CompanyFromContext(ctx),
		EmployeeID:     in.EmployeeID,
		Date:           in.Date,
		FuelType:       in.FuelType,
		Quantity:       in.Quantity,
		UnitCost:       in.UnitCost,
		TotalCost:      in.TotalCost,
		Odometer:       in.Odometer,
		VendorID:       in.VendorID,
		FullTank:       in.FullTank,
		MilesTraveled:  in.MilesTraveled,
		FuelEfficiency: in.FuelEfficiency,
		State:          in.State,
		Reference:      in.Reference,
		Personal:       in.Personal,
		Reset:          in.Reset,
		Latitude:       in.Latitude,
		Longitude:      in.Longitude,
		ExternalID:     in.ExternalID,
		NoSemana:       in.NoSemana,
		EstadoProv:     in.EstadoProv,
		OperatorName:   in.OperatorName,
		UpdatedAt:      now,
	})
	if err != nil {
		return dto.FuelEntryResponse{}, err
	}
	return toFuelEntryResponse(r), nil
}

func (s *FuelEntryStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateFuelEntryRequest) (dto.FuelEntryResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateFuelEntry(ctx, gen.UpdateFuelEntryParams{
		ID:             id,
		ParentID:       parentID,
		CompanyID:      middleware.CompanyFromContext(ctx),
		EmployeeID:     in.EmployeeID,
		Date:           in.Date,
		FuelType:       in.FuelType,
		Quantity:       in.Quantity,
		UnitCost:       in.UnitCost,
		TotalCost:      in.TotalCost,
		Odometer:       in.Odometer,
		VendorID:       in.VendorID,
		FullTank:       in.FullTank,
		MilesTraveled:  in.MilesTraveled,
		FuelEfficiency: in.FuelEfficiency,
		State:          in.State,
		Reference:      in.Reference,
		Personal:       in.Personal,
		Reset:          in.Reset,
		Latitude:       in.Latitude,
		Longitude:      in.Longitude,
		ExternalID:     in.ExternalID,
		UpdatedAt:      now,
		NoSemana:       in.NoSemana,
		EstadoProv:     in.EstadoProv,
		OperatorName:   in.OperatorName,
	})
	if err != nil {
		return dto.FuelEntryResponse{}, err
	}
	return toFuelEntryResponse(r), nil
}

func (s *FuelEntryStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteFuelEntry(ctx, gen.DeleteFuelEntryParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toFuelEntryResponse(r gen.FuelEntry) dto.FuelEntryResponse {
	return dto.FuelEntryResponse{
		ID:             r.ID,
		AssetID:        r.AssetID,
		EmployeeID:     r.EmployeeID,
		Date:           r.Date,
		FuelType:       r.FuelType,
		Quantity:       r.Quantity,
		UnitCost:       r.UnitCost,
		TotalCost:      r.TotalCost,
		Odometer:       r.Odometer,
		VendorID:       r.VendorID,
		FullTank:       r.FullTank,
		MilesTraveled:  r.MilesTraveled,
		FuelEfficiency: r.FuelEfficiency,
		State:          r.State,
		Reference:      r.Reference,
		Personal:       r.Personal,
		Reset:          r.Reset,
		Latitude:       r.Latitude,
		Longitude:      r.Longitude,
		ExternalID:     r.ExternalID,
		UpdatedAt:      r.UpdatedAt,
		NoSemana:       r.NoSemana,
		EstadoProv:     r.EstadoProv,
		OperatorName:   r.OperatorName,
	}
}
