package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type WorkOrderLineItemStore struct{ q *gen.Queries }

func NewWorkOrderLineItemStore(q *gen.Queries) *WorkOrderLineItemStore {
	return &WorkOrderLineItemStore{q: q}
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
	r, err := s.q.CreateWorkOrderLineItem(ctx, gen.CreateWorkOrderLineItemParams{
		ParentID:     parentID,
		CompanyID:    middleware.CompanyFromContext(ctx),
		LineItemType: in.LineItemType,
		Title:        in.Title,
		Description:  in.Description,
		Position:     in.Position,
		ServiceTask:  in.ServiceTask,
		PartsCost:    in.PartsCost,
		LaborCost:    in.LaborCost,
		Subtotal:     in.Subtotal,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return dto.WorkOrderLineItemResponse{}, err
	}
	return toWorkOrderLineItemResponse(r), nil
}

func (s *WorkOrderLineItemStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateWorkOrderLineItemRequest) (dto.WorkOrderLineItemResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateWorkOrderLineItem(ctx, gen.UpdateWorkOrderLineItemParams{
		ID:           id,
		ParentID:     parentID,
		CompanyID:    middleware.CompanyFromContext(ctx),
		LineItemType: in.LineItemType,
		Title:        in.Title,
		Description:  in.Description,
		Position:     in.Position,
		ServiceTask:  in.ServiceTask,
		PartsCost:    in.PartsCost,
		LaborCost:    in.LaborCost,
		Subtotal:     in.Subtotal,
		UpdatedAt:    now,
	})
	if err != nil {
		return dto.WorkOrderLineItemResponse{}, err
	}
	return toWorkOrderLineItemResponse(r), nil
}

func (s *WorkOrderLineItemStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteWorkOrderLineItem(ctx, gen.DeleteWorkOrderLineItemParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
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
