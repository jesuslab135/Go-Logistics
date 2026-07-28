package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type WorkOrderSubLineItemStore struct{ q *gen.Queries }

func NewWorkOrderSubLineItemStore(q *gen.Queries) *WorkOrderSubLineItemStore {
	return &WorkOrderSubLineItemStore{q: q}
}

func (s *WorkOrderSubLineItemStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.WorkOrderSubLineItemResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListWorkOrderSubLineItems(ctx, gen.ListWorkOrderSubLineItemsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountWorkOrderSubLineItems(ctx, gen.CountWorkOrderSubLineItemsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.WorkOrderSubLineItemResponse, len(rows))
	for i, r := range rows {
		out[i] = toWorkOrderSubLineItemResponse(r)
	}
	return out, total, nil
}

func (s *WorkOrderSubLineItemStore) Get(ctx context.Context, parentID, id int64) (dto.WorkOrderSubLineItemResponse, error) {
	r, err := s.q.GetWorkOrderSubLineItem(ctx, gen.GetWorkOrderSubLineItemParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.WorkOrderSubLineItemResponse{}, err
	}
	return toWorkOrderSubLineItemResponse(r), nil
}

func (s *WorkOrderSubLineItemStore) Create(ctx context.Context, parentID int64, in dto.CreateWorkOrderSubLineItemRequest) (dto.WorkOrderSubLineItemResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateWorkOrderSubLineItem(ctx, gen.CreateWorkOrderSubLineItemParams{
		ParentID:             parentID,
		CompanyID:            middleware.CompanyFromContext(ctx),
		ItemType:             in.ItemType,
		Description:          in.Description,
		Position:             in.Position,
		PartID:               in.PartID,
		PartLocationDetailID: in.PartLocationDetailID,
		TechnicianID:         in.TechnicianID,
		UnitCost:             in.UnitCost,
		Quantity:             decimalOrDefault(in.Quantity, 1),
		CreatedAt:            now,
		UpdatedAt:            now,
	})
	if err != nil {
		return dto.WorkOrderSubLineItemResponse{}, err
	}
	return toWorkOrderSubLineItemResponse(r), nil
}

func (s *WorkOrderSubLineItemStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateWorkOrderSubLineItemRequest) (dto.WorkOrderSubLineItemResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateWorkOrderSubLineItem(ctx, gen.UpdateWorkOrderSubLineItemParams{
		ID:                   id,
		ParentID:             parentID,
		CompanyID:            middleware.CompanyFromContext(ctx),
		ItemType:             in.ItemType,
		Description:          in.Description,
		Position:             in.Position,
		PartID:               in.PartID,
		PartLocationDetailID: in.PartLocationDetailID,
		TechnicianID:         in.TechnicianID,
		UnitCost:             in.UnitCost,
		Quantity:             in.Quantity,
		UpdatedAt:            now,
	})
	if err != nil {
		return dto.WorkOrderSubLineItemResponse{}, err
	}
	return toWorkOrderSubLineItemResponse(r), nil
}

func (s *WorkOrderSubLineItemStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteWorkOrderSubLineItem(ctx, gen.DeleteWorkOrderSubLineItemParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toWorkOrderSubLineItemResponse(r gen.WorkOrderSubLineItem) dto.WorkOrderSubLineItemResponse {
	return dto.WorkOrderSubLineItemResponse{
		ID:                   r.ID,
		LineItemID:           r.LineItemID,
		ItemType:             r.ItemType,
		Description:          r.Description,
		Position:             r.Position,
		PartID:               r.PartID,
		PartLocationDetailID: r.PartLocationDetailID,
		TechnicianID:         r.TechnicianID,
		UnitCost:             r.UnitCost,
		Quantity:             r.Quantity,
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
	}
}
