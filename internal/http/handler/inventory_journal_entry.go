package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type InventoryJournalEntryStore struct{ q *gen.Queries }

func NewInventoryJournalEntryStore(q *gen.Queries) *InventoryJournalEntryStore {
	return &InventoryJournalEntryStore{q: q}
}

func (s *InventoryJournalEntryStore) List(ctx context.Context, p paginate.Params) ([]dto.InventoryJournalEntryResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListInventoryJournalEntries(ctx, gen.ListInventoryJournalEntriesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountInventoryJournalEntries(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.InventoryJournalEntryResponse, len(rows))
	for i, r := range rows {
		out[i] = toInventoryJournalEntryResponse(r)
	}
	return out, total, nil
}

func (s *InventoryJournalEntryStore) Get(ctx context.Context, id int64) (dto.InventoryJournalEntryResponse, error) {
	r, err := s.q.GetInventoryJournalEntry(ctx, gen.GetInventoryJournalEntryParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.InventoryJournalEntryResponse{}, err
	}
	return toInventoryJournalEntryResponse(r), nil
}

func (s *InventoryJournalEntryStore) Create(ctx context.Context, in dto.CreateInventoryJournalEntryRequest) (dto.InventoryJournalEntryResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateInventoryJournalEntry(ctx, gen.CreateInventoryJournalEntryParams{
		CompanyID:              middleware.CompanyFromContext(ctx),
		PartID:                 in.PartID,
		PartLocationDetailID:   in.PartLocationDetailID,
		UserID:                 in.UserID,
		PreviousQuantity:       in.PreviousQuantity,
		AdjustmentQuantity:     in.AdjustmentQuantity,
		CurrentQuantity:        in.CurrentQuantity,
		UnitCost:               in.UnitCost,
		ReasonID:               in.ReasonID,
		WorkOrderID:            in.WorkOrderID,
		PurchaseOrderLineID:    in.PurchaseOrderLineID,
		VendorID:               in.VendorID,
		AdjustmentType:         orDefault(in.AdjustmentType, "manual"),
		TransferPartLocationID: in.TransferPartLocationID,
		Notes:                  in.Notes,
		CreatedAt:              now,
	})
	if err != nil {
		return dto.InventoryJournalEntryResponse{}, err
	}
	return toInventoryJournalEntryResponse(r), nil
}

func (s *InventoryJournalEntryStore) Update(ctx context.Context, id int64, in dto.UpdateInventoryJournalEntryRequest) (dto.InventoryJournalEntryResponse, error) {
	r, err := s.q.UpdateInventoryJournalEntry(ctx, gen.UpdateInventoryJournalEntryParams{
		ID:                     id,
		CompanyID:              middleware.CompanyFromContext(ctx),
		PartID:                 in.PartID,
		PartLocationDetailID:   in.PartLocationDetailID,
		UserID:                 in.UserID,
		PreviousQuantity:       in.PreviousQuantity,
		AdjustmentQuantity:     in.AdjustmentQuantity,
		CurrentQuantity:        in.CurrentQuantity,
		UnitCost:               in.UnitCost,
		ReasonID:               in.ReasonID,
		WorkOrderID:            in.WorkOrderID,
		PurchaseOrderLineID:    in.PurchaseOrderLineID,
		VendorID:               in.VendorID,
		AdjustmentType:         in.AdjustmentType,
		TransferPartLocationID: in.TransferPartLocationID,
		Notes:                  in.Notes,
	})
	if err != nil {
		return dto.InventoryJournalEntryResponse{}, err
	}
	return toInventoryJournalEntryResponse(r), nil
}

func (s *InventoryJournalEntryStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteInventoryJournalEntry(ctx, gen.DeleteInventoryJournalEntryParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toInventoryJournalEntryResponse(r gen.InventoryJournalEntry) dto.InventoryJournalEntryResponse {
	return dto.InventoryJournalEntryResponse{
		ID:                     r.ID,
		CompanyID:              r.CompanyID,
		PartID:                 r.PartID,
		PartLocationDetailID:   r.PartLocationDetailID,
		UserID:                 r.UserID,
		PreviousQuantity:       r.PreviousQuantity,
		AdjustmentQuantity:     r.AdjustmentQuantity,
		CurrentQuantity:        r.CurrentQuantity,
		UnitCost:               r.UnitCost,
		ReasonID:               r.ReasonID,
		WorkOrderID:            r.WorkOrderID,
		PurchaseOrderLineID:    r.PurchaseOrderLineID,
		VendorID:               r.VendorID,
		AdjustmentType:         r.AdjustmentType,
		TransferPartLocationID: r.TransferPartLocationID,
		Notes:                  r.Notes,
		CreatedAt:              r.CreatedAt,
	}
}
