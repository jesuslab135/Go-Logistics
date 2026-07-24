package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type PurchaseOrderLineItemStore struct{ q *gen.Queries }

func NewPurchaseOrderLineItemStore(q *gen.Queries) *PurchaseOrderLineItemStore {
	return &PurchaseOrderLineItemStore{q: q}
}

func (s *PurchaseOrderLineItemStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.PurchaseOrderLineItemResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListPurchaseOrderLineItems(ctx, gen.ListPurchaseOrderLineItemsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountPurchaseOrderLineItems(ctx, gen.CountPurchaseOrderLineItemsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PurchaseOrderLineItemResponse, len(rows))
	for i, r := range rows {
		out[i] = toPurchaseOrderLineItemResponse(r)
	}
	return out, total, nil
}

func (s *PurchaseOrderLineItemStore) Get(ctx context.Context, parentID, id int64) (dto.PurchaseOrderLineItemResponse, error) {
	r, err := s.q.GetPurchaseOrderLineItem(ctx, gen.GetPurchaseOrderLineItemParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.PurchaseOrderLineItemResponse{}, err
	}
	return toPurchaseOrderLineItemResponse(r), nil
}

func (s *PurchaseOrderLineItemStore) Create(ctx context.Context, parentID int64, in dto.CreatePurchaseOrderLineItemRequest) (dto.PurchaseOrderLineItemResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreatePurchaseOrderLineItem(ctx, gen.CreatePurchaseOrderLineItemParams{
		ParentID:      parentID,
		CompanyID:     middleware.CompanyFromContext(ctx),
		PartID:        in.PartID,
		Quantity:      in.Quantity,
		TotalReceived: in.TotalReceived,
		UnitCost:      in.UnitCost,
		Subtotal:      in.Subtotal,
		Position:      in.Position,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		return dto.PurchaseOrderLineItemResponse{}, err
	}
	return toPurchaseOrderLineItemResponse(r), nil
}

func (s *PurchaseOrderLineItemStore) Update(ctx context.Context, parentID, id int64, in dto.UpdatePurchaseOrderLineItemRequest) (dto.PurchaseOrderLineItemResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdatePurchaseOrderLineItem(ctx, gen.UpdatePurchaseOrderLineItemParams{
		ID:            id,
		ParentID:      parentID,
		CompanyID:     middleware.CompanyFromContext(ctx),
		PartID:        in.PartID,
		Quantity:      in.Quantity,
		TotalReceived: in.TotalReceived,
		UnitCost:      in.UnitCost,
		Subtotal:      in.Subtotal,
		Position:      in.Position,
		UpdatedAt:     now,
	})
	if err != nil {
		return dto.PurchaseOrderLineItemResponse{}, err
	}
	return toPurchaseOrderLineItemResponse(r), nil
}

func (s *PurchaseOrderLineItemStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeletePurchaseOrderLineItem(ctx, gen.DeletePurchaseOrderLineItemParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toPurchaseOrderLineItemResponse(r gen.PurchaseOrderLineItem) dto.PurchaseOrderLineItemResponse {
	return dto.PurchaseOrderLineItemResponse{
		ID:              r.ID,
		PurchaseOrderID: r.PurchaseOrderID,
		PartID:          r.PartID,
		Quantity:        r.Quantity,
		TotalReceived:   r.TotalReceived,
		UnitCost:        r.UnitCost,
		Subtotal:        r.Subtotal,
		Position:        r.Position,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}
