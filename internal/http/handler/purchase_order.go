package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

// Totals are derived from the line items and this record's own rate terms,
// so a write that can change either recomputes them in the same transaction.
type PurchaseOrderStore struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewPurchaseOrderStore(q *gen.Queries, pool *pgxpool.Pool) *PurchaseOrderStore {
	return &PurchaseOrderStore{q: q, pool: pool}
}

func (s *PurchaseOrderStore) List(ctx context.Context, p paginate.Params) ([]dto.PurchaseOrderResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListPurchaseOrders(ctx, gen.ListPurchaseOrdersParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountPurchaseOrders(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PurchaseOrderResponse, len(rows))
	for i, r := range rows {
		out[i] = toPurchaseOrderResponse(r)
	}
	return out, total, nil
}

func (s *PurchaseOrderStore) Get(ctx context.Context, id int64) (dto.PurchaseOrderResponse, error) {
	r, err := s.q.GetPurchaseOrder(ctx, gen.GetPurchaseOrderParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.PurchaseOrderResponse{}, err
	}
	return toPurchaseOrderResponse(r), nil
}

func (s *PurchaseOrderStore) Create(ctx context.Context, in dto.CreatePurchaseOrderRequest) (dto.PurchaseOrderResponse, error) {
	var out dto.PurchaseOrderResponse
	err := inTx(ctx, s.pool, s.q, func(qtx *gen.Queries) error {
		now := time.Now().UTC()
		r, err := qtx.CreatePurchaseOrder(ctx, gen.CreatePurchaseOrderParams{
			CompanyID:          middleware.CompanyFromContext(ctx),
			CreatedByID:        authorFromContext(ctx),
			Number:             in.Number,
			Description:        in.Description,
			VendorID:           in.VendorID,
			DestinationID:      in.DestinationID,
			DiscountType:       orDefault(in.DiscountType, "FIXED"),
			Discount:           in.Discount,
			DiscountPercentage: in.DiscountPercentage,
			Tax1Type:           orDefault(in.Tax1Type, "PERCENTAGE"),
			Tax1:               in.Tax1,
			Tax1Percentage:     in.Tax1Percentage,
			Tax2Type:           orDefault(in.Tax2Type, "PERCENTAGE"),
			Tax2:               in.Tax2,
			Tax2Percentage:     in.Tax2Percentage,
			Shipping:           in.Shipping,
			Labels:             jsonbOrDefault(in.Labels, "[]"),
			CustomFields:       jsonbOrDefault(in.CustomFields, "{}"),
			CreatedAt:          now,
			UpdatedAt:          now,
		})
		if err != nil {
			return err
		}
		if err := recalcPurchaseOrder(ctx, qtx, r.ID, now); err != nil {
			return err
		}
		fresh, err := qtx.GetPurchaseOrder(ctx, gen.GetPurchaseOrderParams{ID: r.ID, CompanyID: middleware.CompanyFromContext(ctx)})
		if err != nil {
			return err
		}
		out = toPurchaseOrderResponse(fresh)
		return nil
	})
	return out, err
}

func (s *PurchaseOrderStore) Update(ctx context.Context, id int64, in dto.UpdatePurchaseOrderRequest) (dto.PurchaseOrderResponse, error) {
	var out dto.PurchaseOrderResponse
	err := inTx(ctx, s.pool, s.q, func(qtx *gen.Queries) error {
		now := time.Now().UTC()
		_, err := qtx.UpdatePurchaseOrder(ctx, gen.UpdatePurchaseOrderParams{
			ID:                 id,
			CompanyID:          middleware.CompanyFromContext(ctx),
			Number:             in.Number,
			Description:        in.Description,
			VendorID:           in.VendorID,
			DestinationID:      in.DestinationID,
			DiscountType:       in.DiscountType,
			Discount:           in.Discount,
			DiscountPercentage: in.DiscountPercentage,
			Tax1Type:           in.Tax1Type,
			Tax1:               in.Tax1,
			Tax1Percentage:     in.Tax1Percentage,
			Tax2Type:           in.Tax2Type,
			Tax2:               in.Tax2,
			Tax2Percentage:     in.Tax2Percentage,
			Shipping:           in.Shipping,
			Labels:             jsonbOrDefault(in.Labels, "[]"),
			CustomFields:       jsonbOrDefault(in.CustomFields, "{}"),
			UpdatedAt:          now,
		})
		if err != nil {
			return err
		}
		if err := recalcPurchaseOrder(ctx, qtx, id, now); err != nil {
			return err
		}
		fresh, err := qtx.GetPurchaseOrder(ctx, gen.GetPurchaseOrderParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
		if err != nil {
			return err
		}
		out = toPurchaseOrderResponse(fresh)
		return nil
	})
	return out, err
}

func (s *PurchaseOrderStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeletePurchaseOrder(ctx, gen.DeletePurchaseOrderParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toPurchaseOrderResponse(r gen.PurchaseOrder) dto.PurchaseOrderResponse {
	return dto.PurchaseOrderResponse{
		ID:                 r.ID,
		CompanyID:          r.CompanyID,
		Number:             r.Number,
		Description:        r.Description,
		State:              r.State,
		VendorID:           r.VendorID,
		DestinationID:      r.DestinationID,
		DiscountType:       r.DiscountType,
		Discount:           r.Discount,
		DiscountPercentage: r.DiscountPercentage,
		Tax1Type:           r.Tax1Type,
		Tax1:               r.Tax1,
		Tax1Percentage:     r.Tax1Percentage,
		Tax2Type:           r.Tax2Type,
		Tax2:               r.Tax2,
		Tax2Percentage:     r.Tax2Percentage,
		Shipping:           r.Shipping,
		Subtotal:           r.Subtotal,
		TotalAmount:        r.TotalAmount,
		CreatedByID:        r.CreatedByID,
		SubmittedAt:        r.SubmittedAt,
		SubmittedByID:      r.SubmittedByID,
		RejectedAt:         r.RejectedAt,
		RejectedByID:       r.RejectedByID,
		ApprovedAt:         r.ApprovedAt,
		ApprovedByID:       r.ApprovedByID,
		PurchasedAt:        r.PurchasedAt,
		ReceivedPartialAt:  r.ReceivedPartialAt,
		ReceivedFullAt:     r.ReceivedFullAt,
		ClosedAt:           r.ClosedAt,
		Labels:             json.RawMessage(r.Labels),
		CustomFields:       json.RawMessage(r.CustomFields),
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}
}
