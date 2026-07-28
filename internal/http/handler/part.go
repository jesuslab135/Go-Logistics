package handler

import (
	"context"
	"encoding/json"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type PartStore struct{ q *gen.Queries }

func NewPartStore(q *gen.Queries) *PartStore { return &PartStore{q: q} }

func (s *PartStore) List(ctx context.Context, p paginate.Params) ([]dto.PartResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListParts(ctx, gen.ListPartsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountParts(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PartResponse, len(rows))
	for i, r := range rows {
		out[i] = toPartResponse(r)
	}
	return out, total, nil
}

func (s *PartStore) Get(ctx context.Context, id int64) (dto.PartResponse, error) {
	r, err := s.q.GetPart(ctx, gen.GetPartParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.PartResponse{}, err
	}
	return toPartResponse(r), nil
}

func (s *PartStore) Create(ctx context.Context, in dto.CreatePartRequest) (dto.PartResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreatePart(ctx, gen.CreatePartParams{
		CompanyID:              middleware.CompanyFromContext(ctx),
		PartNumber:             in.PartNumber,
		Description:            in.Description,
		UsefulLifeMonths:       in.UsefulLifeMonths,
		UsefulLifeDistance:     in.UsefulLifeDistance,
		PartCategoryID:         in.PartCategoryID,
		PartManufacturerID:     in.PartManufacturerID,
		MeasurementUnitID:      in.MeasurementUnitID,
		ManufacturerPartNumber: in.ManufacturerPartNumber,
		SupplierPartNumber:     in.SupplierPartNumber,
		Upc:                    in.Upc,
		UnitCost:               in.UnitCost,
		InventoryItem:          boolOrDefault(in.InventoryItem, true),
		ArchivedAt:             in.ArchivedAt,
		CustomFields:           jsonbOrDefault(in.CustomFields, "{}"),
		CreatedAt:              now,
		UpdatedAt:              now,
	})
	if err != nil {
		return dto.PartResponse{}, err
	}
	return toPartResponse(r), nil
}

func (s *PartStore) Update(ctx context.Context, id int64, in dto.UpdatePartRequest) (dto.PartResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdatePart(ctx, gen.UpdatePartParams{
		ID:                     id,
		CompanyID:              middleware.CompanyFromContext(ctx),
		PartNumber:             in.PartNumber,
		Description:            in.Description,
		UsefulLifeMonths:       in.UsefulLifeMonths,
		UsefulLifeDistance:     in.UsefulLifeDistance,
		PartCategoryID:         in.PartCategoryID,
		PartManufacturerID:     in.PartManufacturerID,
		MeasurementUnitID:      in.MeasurementUnitID,
		ManufacturerPartNumber: in.ManufacturerPartNumber,
		SupplierPartNumber:     in.SupplierPartNumber,
		Upc:                    in.Upc,
		UnitCost:               in.UnitCost,
		InventoryItem:          in.InventoryItem,
		ArchivedAt:             in.ArchivedAt,
		CustomFields:           jsonbOrDefault(in.CustomFields, "{}"),
		UpdatedAt:              now,
	})
	if err != nil {
		return dto.PartResponse{}, err
	}
	return toPartResponse(r), nil
}

func (s *PartStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeletePart(ctx, gen.DeletePartParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toPartResponse(r gen.Part) dto.PartResponse {
	return dto.PartResponse{
		ID:                     r.ID,
		CompanyID:              r.CompanyID,
		PartNumber:             r.PartNumber,
		Description:            r.Description,
		UsefulLifeMonths:       r.UsefulLifeMonths,
		UsefulLifeDistance:     r.UsefulLifeDistance,
		PartCategoryID:         r.PartCategoryID,
		PartManufacturerID:     r.PartManufacturerID,
		MeasurementUnitID:      r.MeasurementUnitID,
		ManufacturerPartNumber: r.ManufacturerPartNumber,
		SupplierPartNumber:     r.SupplierPartNumber,
		Upc:                    r.Upc,
		UnitCost:               r.UnitCost,
		InventoryItem:          r.InventoryItem,
		ArchivedAt:             r.ArchivedAt,
		CustomFields:           json.RawMessage(r.CustomFields),
		CreatedAt:              r.CreatedAt,
		UpdatedAt:              r.UpdatedAt,
	}
}
