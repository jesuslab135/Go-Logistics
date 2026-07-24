package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type PartInventoryStore struct{ q *gen.Queries }

func NewPartInventoryStore(q *gen.Queries) *PartInventoryStore { return &PartInventoryStore{q: q} }

func (s *PartInventoryStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.PartInventoryResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListPartInventories(ctx, gen.ListPartInventoriesParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountPartInventories(ctx, gen.CountPartInventoriesParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PartInventoryResponse, len(rows))
	for i, r := range rows {
		out[i] = toPartInventoryResponse(r)
	}
	return out, total, nil
}

func (s *PartInventoryStore) Get(ctx context.Context, parentID, id int64) (dto.PartInventoryResponse, error) {
	r, err := s.q.GetPartInventory(ctx, gen.GetPartInventoryParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.PartInventoryResponse{}, err
	}
	return toPartInventoryResponse(r), nil
}

func (s *PartInventoryStore) Create(ctx context.Context, parentID int64, in dto.CreatePartInventoryRequest) (dto.PartInventoryResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreatePartInventory(ctx, gen.CreatePartInventoryParams{
		ParentID:                   parentID,
		CompanyID:                  middleware.CompanyFromContext(ctx),
		LocationID:                 in.LocationID,
		AvailableQuantity:          in.AvailableQuantity,
		ExpiryDate:                 in.ExpiryDate,
		Aisle:                      in.Aisle,
		Row:                        in.Row,
		Bin:                        in.Bin,
		ReorderPoint:               in.ReorderPoint,
		ReorderPointEnabled:        in.ReorderPointEnabled,
		ReorderQuantity:            in.ReorderQuantity,
		ReorderPointLeadTimeDays:   in.ReorderPointLeadTimeDays,
		Active:                     in.Active,
		TrackInventory:             in.TrackInventory,
		AverageUnitCost:            in.AverageUnitCost,
		AvailableQuantityUpdatedAt: in.AvailableQuantityUpdatedAt,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	})
	if err != nil {
		return dto.PartInventoryResponse{}, err
	}
	return toPartInventoryResponse(r), nil
}

func (s *PartInventoryStore) Update(ctx context.Context, parentID, id int64, in dto.UpdatePartInventoryRequest) (dto.PartInventoryResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdatePartInventory(ctx, gen.UpdatePartInventoryParams{
		ID:                         id,
		ParentID:                   parentID,
		CompanyID:                  middleware.CompanyFromContext(ctx),
		LocationID:                 in.LocationID,
		AvailableQuantity:          in.AvailableQuantity,
		ExpiryDate:                 in.ExpiryDate,
		Aisle:                      in.Aisle,
		Row:                        in.Row,
		Bin:                        in.Bin,
		ReorderPoint:               in.ReorderPoint,
		ReorderPointEnabled:        in.ReorderPointEnabled,
		ReorderQuantity:            in.ReorderQuantity,
		ReorderPointLeadTimeDays:   in.ReorderPointLeadTimeDays,
		Active:                     in.Active,
		TrackInventory:             in.TrackInventory,
		AverageUnitCost:            in.AverageUnitCost,
		AvailableQuantityUpdatedAt: in.AvailableQuantityUpdatedAt,
		UpdatedAt:                  now,
	})
	if err != nil {
		return dto.PartInventoryResponse{}, err
	}
	return toPartInventoryResponse(r), nil
}

func (s *PartInventoryStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeletePartInventory(ctx, gen.DeletePartInventoryParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toPartInventoryResponse(r gen.PartInventory) dto.PartInventoryResponse {
	return dto.PartInventoryResponse{
		ID:                         r.ID,
		PartID:                     r.PartID,
		LocationID:                 r.LocationID,
		AvailableQuantity:          r.AvailableQuantity,
		ExpiryDate:                 r.ExpiryDate,
		Aisle:                      r.Aisle,
		Row:                        r.Row,
		Bin:                        r.Bin,
		ReorderPoint:               r.ReorderPoint,
		ReorderPointEnabled:        r.ReorderPointEnabled,
		ReorderQuantity:            r.ReorderQuantity,
		ReorderPointLeadTimeDays:   r.ReorderPointLeadTimeDays,
		Active:                     r.Active,
		TrackInventory:             r.TrackInventory,
		AverageUnitCost:            r.AverageUnitCost,
		AvailableQuantityUpdatedAt: r.AvailableQuantityUpdatedAt,
		CreatedAt:                  r.CreatedAt,
		UpdatedAt:                  r.UpdatedAt,
	}
}
