package handler

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

// A line prices itself (quantity x unit_cost) and the order sums its lines, so
// every write recomputes both in the same transaction.
type PurchaseOrderLineItemStore struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewPurchaseOrderLineItemStore(q *gen.Queries, pool *pgxpool.Pool) *PurchaseOrderLineItemStore {
	return &PurchaseOrderLineItemStore{q: q, pool: pool}
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
	company := middleware.CompanyFromContext(ctx)

	var out dto.PurchaseOrderLineItemResponse
	err := inTx(ctx, s.pool, s.q, func(qtx *gen.Queries) error {
		r, err := qtx.CreatePurchaseOrderLineItem(ctx, gen.CreatePurchaseOrderLineItemParams{
			ParentID:      parentID,
			CompanyID:     company,
			PartID:        in.PartID,
			Quantity:      in.Quantity,
			TotalReceived: in.TotalReceived,
			UnitCost:      in.UnitCost,
			Position:      in.Position,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
		if err != nil {
			return err
		}
		if err := recalcPurchaseOrderLineItem(ctx, qtx, r.ID, now); err != nil {
			return err
		}
		// Re-read: the line's subtotal was just derived.
		r, err = qtx.GetPurchaseOrderLineItem(ctx, gen.GetPurchaseOrderLineItemParams{ID: r.ID, ParentID: parentID, CompanyID: company})
		if err != nil {
			return err
		}
		out = toPurchaseOrderLineItemResponse(r)
		return nil
	})
	return out, err
}

func (s *PurchaseOrderLineItemStore) Update(ctx context.Context, parentID, id int64, in dto.UpdatePurchaseOrderLineItemRequest) (dto.PurchaseOrderLineItemResponse, error) {
	now := time.Now().UTC()
	company := middleware.CompanyFromContext(ctx)

	var out dto.PurchaseOrderLineItemResponse
	err := inTx(ctx, s.pool, s.q, func(qtx *gen.Queries) error {
		r, err := qtx.UpdatePurchaseOrderLineItem(ctx, gen.UpdatePurchaseOrderLineItemParams{
			ID:            id,
			ParentID:      parentID,
			CompanyID:     company,
			PartID:        in.PartID,
			Quantity:      in.Quantity,
			TotalReceived: in.TotalReceived,
			UnitCost:      in.UnitCost,
			Position:      in.Position,
			UpdatedAt:     now,
		})
		if err != nil {
			return err
		}
		if err := recalcPurchaseOrderLineItem(ctx, qtx, r.ID, now); err != nil {
			return err
		}
		r, err = qtx.GetPurchaseOrderLineItem(ctx, gen.GetPurchaseOrderLineItemParams{ID: id, ParentID: parentID, CompanyID: company})
		if err != nil {
			return err
		}
		out = toPurchaseOrderLineItemResponse(r)
		return nil
	})
	return out, err
}

func (s *PurchaseOrderLineItemStore) Delete(ctx context.Context, parentID, id int64) error {
	now := time.Now().UTC()
	return inTx(ctx, s.pool, s.q, func(qtx *gen.Queries) error {
		if err := qtx.DeletePurchaseOrderLineItem(ctx, gen.DeletePurchaseOrderLineItemParams{
			ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx),
		}); err != nil {
			return err
		}
		return recalcPurchaseOrder(ctx, qtx, parentID, now)
	})
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
