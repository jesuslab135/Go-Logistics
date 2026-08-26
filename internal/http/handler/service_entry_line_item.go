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

// A line's subtotal is derived (quantity x unit_cost) and the entry sums its
// lines, so every write recomputes both in the same transaction.
//
// parts_cost and labor_cost stay inputs: they are the operator's split of the
// line between the two buckets, which the schema gives every line and which
// nothing else can infer. The entry's parts_subtotal and labor_subtotal sum them.
type ServiceEntryLineItemStore struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewServiceEntryLineItemStore(q *gen.Queries, pool *pgxpool.Pool) *ServiceEntryLineItemStore {
	return &ServiceEntryLineItemStore{q: q, pool: pool}
}

func (s *ServiceEntryLineItemStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.ServiceEntryLineItemResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListServiceEntryLineItems(ctx, gen.ListServiceEntryLineItemsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountServiceEntryLineItems(ctx, gen.CountServiceEntryLineItemsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.ServiceEntryLineItemResponse, len(rows))
	for i, r := range rows {
		out[i] = toServiceEntryLineItemResponse(r)
	}
	return out, total, nil
}

func (s *ServiceEntryLineItemStore) Get(ctx context.Context, parentID, id int64) (dto.ServiceEntryLineItemResponse, error) {
	r, err := s.q.GetServiceEntryLineItem(ctx, gen.GetServiceEntryLineItemParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.ServiceEntryLineItemResponse{}, err
	}
	return toServiceEntryLineItemResponse(r), nil
}

func (s *ServiceEntryLineItemStore) Create(ctx context.Context, parentID int64, in dto.CreateServiceEntryLineItemRequest) (dto.ServiceEntryLineItemResponse, error) {
	now := time.Now().UTC()
	company := middleware.CompanyFromContext(ctx)

	var out dto.ServiceEntryLineItemResponse
	err := inTx(ctx, s.pool, s.q, func(qtx *gen.Queries) error {
		r, err := qtx.CreateServiceEntryLineItem(ctx, gen.CreateServiceEntryLineItemParams{
			ParentID:          parentID,
			CompanyID:         company,
			LineItemType:      in.LineItemType,
			Description:       in.Description,
			ServiceTaskID:     in.ServiceTaskID,
			PartID:            in.PartID,
			TechnicianID:      in.TechnicianID,
			TireID:            in.TireID,
			ServiceReminderID: in.ServiceReminderID,
			UnitCost:          in.UnitCost,
			Quantity:          decimalOrDefault(in.Quantity, 1),
			PartsCost:         in.PartsCost,
			LaborCost:         in.LaborCost,
			Position:          in.Position,
			CreatedAt:         now,
			UpdatedAt:         now,
		})
		if err != nil {
			return err
		}
		if err := recalcServiceEntryLineItem(ctx, qtx, r.ID, now); err != nil {
			return err
		}
		r, err = qtx.GetServiceEntryLineItem(ctx, gen.GetServiceEntryLineItemParams{ID: r.ID, ParentID: parentID, CompanyID: company})
		if err != nil {
			return err
		}
		out = toServiceEntryLineItemResponse(r)
		return nil
	})
	return out, err
}

func (s *ServiceEntryLineItemStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateServiceEntryLineItemRequest) (dto.ServiceEntryLineItemResponse, error) {
	now := time.Now().UTC()
	company := middleware.CompanyFromContext(ctx)

	var out dto.ServiceEntryLineItemResponse
	err := inTx(ctx, s.pool, s.q, func(qtx *gen.Queries) error {
		r, err := qtx.UpdateServiceEntryLineItem(ctx, gen.UpdateServiceEntryLineItemParams{
			ID:                id,
			ParentID:          parentID,
			CompanyID:         company,
			LineItemType:      in.LineItemType,
			Description:       in.Description,
			ServiceTaskID:     in.ServiceTaskID,
			PartID:            in.PartID,
			TechnicianID:      in.TechnicianID,
			TireID:            in.TireID,
			ServiceReminderID: in.ServiceReminderID,
			UnitCost:          in.UnitCost,
			Quantity:          in.Quantity,
			PartsCost:         in.PartsCost,
			LaborCost:         in.LaborCost,
			Position:          in.Position,
			UpdatedAt:         now,
		})
		if err != nil {
			return err
		}
		if err := recalcServiceEntryLineItem(ctx, qtx, r.ID, now); err != nil {
			return err
		}
		r, err = qtx.GetServiceEntryLineItem(ctx, gen.GetServiceEntryLineItemParams{ID: id, ParentID: parentID, CompanyID: company})
		if err != nil {
			return err
		}
		out = toServiceEntryLineItemResponse(r)
		return nil
	})
	return out, err
}

func (s *ServiceEntryLineItemStore) Delete(ctx context.Context, parentID, id int64) error {
	now := time.Now().UTC()
	return inTx(ctx, s.pool, s.q, func(qtx *gen.Queries) error {
		if err := qtx.DeleteServiceEntryLineItem(ctx, gen.DeleteServiceEntryLineItemParams{
			ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx),
		}); err != nil {
			return err
		}
		return recalcServiceEntry(ctx, qtx, parentID, now)
	})
}

func toServiceEntryLineItemResponse(r gen.ServiceEntryLineItem) dto.ServiceEntryLineItemResponse {
	return dto.ServiceEntryLineItemResponse{
		ID:                r.ID,
		ServiceEntryID:    r.ServiceEntryID,
		LineItemType:      r.LineItemType,
		Description:       r.Description,
		ServiceTaskID:     r.ServiceTaskID,
		PartID:            r.PartID,
		TechnicianID:      r.TechnicianID,
		TireID:            r.TireID,
		ServiceReminderID: r.ServiceReminderID,
		UnitCost:          r.UnitCost,
		Quantity:          r.Quantity,
		PartsCost:         r.PartsCost,
		LaborCost:         r.LaborCost,
		Subtotal:          r.Subtotal,
		Position:          r.Position,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}
