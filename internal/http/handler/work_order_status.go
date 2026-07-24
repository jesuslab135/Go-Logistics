package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type WorkOrderStatusStore struct{ q *gen.Queries }

func NewWorkOrderStatusStore(q *gen.Queries) *WorkOrderStatusStore {
	return &WorkOrderStatusStore{q: q}
}

func (s *WorkOrderStatusStore) List(ctx context.Context, p paginate.Params) ([]dto.WorkOrderStatusResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListWorkOrderStatuses(ctx, gen.ListWorkOrderStatusesParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountWorkOrderStatuses(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.WorkOrderStatusResponse, len(rows))
	for i, r := range rows {
		out[i] = toWorkOrderStatusResponse(r)
	}
	return out, total, nil
}

func (s *WorkOrderStatusStore) Get(ctx context.Context, id int64) (dto.WorkOrderStatusResponse, error) {
	r, err := s.q.GetWorkOrderStatus(ctx, gen.GetWorkOrderStatusParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.WorkOrderStatusResponse{}, err
	}
	return toWorkOrderStatusResponse(r), nil
}

func (s *WorkOrderStatusStore) Create(ctx context.Context, in dto.CreateWorkOrderStatusRequest) (dto.WorkOrderStatusResponse, error) {
	r, err := s.q.CreateWorkOrderStatus(ctx, gen.CreateWorkOrderStatusParams{
		CompanyID:        middleware.CompanyFromContext(ctx),
		Name:             in.Name,
		Description:      in.Description,
		Color:            in.Color,
		IsDefault:        in.IsDefault,
		MarksAsCompleted: in.MarksAsCompleted,
		Position:         in.Position,
	})
	if err != nil {
		return dto.WorkOrderStatusResponse{}, err
	}
	return toWorkOrderStatusResponse(r), nil
}

func (s *WorkOrderStatusStore) Update(ctx context.Context, id int64, in dto.UpdateWorkOrderStatusRequest) (dto.WorkOrderStatusResponse, error) {
	r, err := s.q.UpdateWorkOrderStatus(ctx, gen.UpdateWorkOrderStatusParams{
		ID:               id,
		CompanyID:        middleware.CompanyFromContext(ctx),
		Name:             in.Name,
		Description:      in.Description,
		Color:            in.Color,
		IsDefault:        in.IsDefault,
		MarksAsCompleted: in.MarksAsCompleted,
		Position:         in.Position,
	})
	if err != nil {
		return dto.WorkOrderStatusResponse{}, err
	}
	return toWorkOrderStatusResponse(r), nil
}

func (s *WorkOrderStatusStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteWorkOrderStatus(ctx, gen.DeleteWorkOrderStatusParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toWorkOrderStatusResponse(r gen.WorkOrderStatus) dto.WorkOrderStatusResponse {
	return dto.WorkOrderStatusResponse{
		ID:               r.ID,
		CompanyID:        r.CompanyID,
		Name:             r.Name,
		Description:      r.Description,
		Color:            r.Color,
		IsDefault:        r.IsDefault,
		MarksAsCompleted: r.MarksAsCompleted,
		Position:         r.Position,
	}
}
