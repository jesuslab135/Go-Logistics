package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type ServiceEntryLineItemStore struct{ q *gen.Queries }

func NewServiceEntryLineItemStore(q *gen.Queries) *ServiceEntryLineItemStore {
	return &ServiceEntryLineItemStore{q: q}
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
	r, err := s.q.CreateServiceEntryLineItem(ctx, gen.CreateServiceEntryLineItemParams{
		ParentID:          parentID,
		CompanyID:         middleware.CompanyFromContext(ctx),
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
		Subtotal:          in.Subtotal,
		Position:          in.Position,
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		return dto.ServiceEntryLineItemResponse{}, err
	}
	return toServiceEntryLineItemResponse(r), nil
}

func (s *ServiceEntryLineItemStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateServiceEntryLineItemRequest) (dto.ServiceEntryLineItemResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateServiceEntryLineItem(ctx, gen.UpdateServiceEntryLineItemParams{
		ID:                id,
		ParentID:          parentID,
		CompanyID:         middleware.CompanyFromContext(ctx),
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
		Subtotal:          in.Subtotal,
		Position:          in.Position,
		UpdatedAt:         now,
	})
	if err != nil {
		return dto.ServiceEntryLineItemResponse{}, err
	}
	return toServiceEntryLineItemResponse(r), nil
}

func (s *ServiceEntryLineItemStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteServiceEntryLineItem(ctx, gen.DeleteServiceEntryLineItemParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
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
