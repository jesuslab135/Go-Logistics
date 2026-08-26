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

// A line's parts_cost, labor_cost and subtotal are derived from its sub-line
// items, and the work order's totals from all of them, so every write here
// recomputes both levels in the same transaction.
type WorkOrderLineItemStore struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewWorkOrderLineItemStore(q *gen.Queries, pool *pgxpool.Pool) *WorkOrderLineItemStore {
	return &WorkOrderLineItemStore{q: q, pool: pool}
}

func (s *WorkOrderLineItemStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.WorkOrderLineItemResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListWorkOrderLineItems(ctx, gen.ListWorkOrderLineItemsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountWorkOrderLineItems(ctx, gen.CountWorkOrderLineItemsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.WorkOrderLineItemResponse, len(rows))
	for i, r := range rows {
		out[i] = toWorkOrderLineItemResponse(r)
	}
	return out, total, nil
}

func (s *WorkOrderLineItemStore) Get(ctx context.Context, parentID, id int64) (dto.WorkOrderLineItemResponse, error) {
	r, err := s.q.GetWorkOrderLineItem(ctx, gen.GetWorkOrderLineItemParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.WorkOrderLineItemResponse{}, err
	}
	return toWorkOrderLineItemResponse(r), nil
}

func (s *WorkOrderLineItemStore) Create(ctx context.Context, parentID int64, in dto.CreateWorkOrderLineItemRequest) (dto.WorkOrderLineItemResponse, error) {
	now := time.Now().UTC()
	company := middleware.CompanyFromContext(ctx)

	var out dto.WorkOrderLineItemResponse
	err := inTx(ctx, s.pool, s.q, func(qtx *gen.Queries) error {
		r, err := qtx.CreateWorkOrderLineItem(ctx, gen.CreateWorkOrderLineItemParams{
			ParentID:     parentID,
			CompanyID:    company,
			LineItemType: in.LineItemType,
			Title:        in.Title,
			Description:  in.Description,
			Position:     in.Position,
			ServiceTask:  in.ServiceTask,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
		if err != nil {
			return err
		}
		// A new line has no sub-line items yet, so its own costs are zero; the
		// document is still recomputed because an empty line changes nothing but
		// the caller should see consistent numbers either way.
		if err := recalcWorkOrderLineItem(ctx, qtx, r.ID, now); err != nil {
			return err
		}
		r, err = qtx.GetWorkOrderLineItem(ctx, gen.GetWorkOrderLineItemParams{ID: r.ID, ParentID: parentID, CompanyID: company})
		if err != nil {
			return err
		}
		out = toWorkOrderLineItemResponse(r)
		return nil
	})
	return out, err
}

func (s *WorkOrderLineItemStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateWorkOrderLineItemRequest) (dto.WorkOrderLineItemResponse, error) {
	now := time.Now().UTC()
	company := middleware.CompanyFromContext(ctx)

	var out dto.WorkOrderLineItemResponse
	err := inTx(ctx, s.pool, s.q, func(qtx *gen.Queries) error {
		r, err := qtx.UpdateWorkOrderLineItem(ctx, gen.UpdateWorkOrderLineItemParams{
			ID:           id,
			ParentID:     parentID,
			CompanyID:    company,
			LineItemType: in.LineItemType,
			Title:        in.Title,
			Description:  in.Description,
			Position:     in.Position,
			ServiceTask:  in.ServiceTask,
			UpdatedAt:    now,
		})
		if err != nil {
			return err
		}
		if err := recalcWorkOrderLineItem(ctx, qtx, r.ID, now); err != nil {
			return err
		}
		r, err = qtx.GetWorkOrderLineItem(ctx, gen.GetWorkOrderLineItemParams{ID: id, ParentID: parentID, CompanyID: company})
		if err != nil {
			return err
		}
		out = toWorkOrderLineItemResponse(r)
		return nil
	})
	return out, err
}

func (s *WorkOrderLineItemStore) Delete(ctx context.Context, parentID, id int64) error {
	now := time.Now().UTC()
	return inTx(ctx, s.pool, s.q, func(qtx *gen.Queries) error {
		if err := qtx.DeleteWorkOrderLineItem(ctx, gen.DeleteWorkOrderLineItemParams{
			ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx),
		}); err != nil {
			return err
		}
		// The line is gone, so only the document is recomputed - and it must be,
		// or its total still includes work that no longer exists.
		return recalcWorkOrder(ctx, qtx, parentID, now)
	})
}

func toWorkOrderLineItemResponse(r gen.WorkOrderLineItem) dto.WorkOrderLineItemResponse {
	return dto.WorkOrderLineItemResponse{
		ID:           r.ID,
		WorkOrderID:  r.WorkOrderID,
		LineItemType: r.LineItemType,
		Title:        r.Title,
		Description:  r.Description,
		Position:     r.Position,
		ServiceTask:  r.ServiceTask,
		PartsCost:    r.PartsCost,
		LaborCost:    r.LaborCost,
		Subtotal:     r.Subtotal,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}
