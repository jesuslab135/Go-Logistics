package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

// WorkOrderStatusLogStore reads a work order's status history. It has no write
// methods by design: rows are appended by whatever changes the status, inside
// that operation's transaction (see logStatusChange in work_order.go).
//
// This restores what the Django model documented — "Historial inmutable de
// transiciones... Se alimenta automáticamente mediante signals" — which the port
// lost by registering the table as a writable nested resource. While it was
// writable, a status could change with no row recorded, and a recorded row could
// be edited to say something that never happened.
type WorkOrderStatusLogStore struct{ q *gen.Queries }

func NewWorkOrderStatusLogStore(q *gen.Queries) *WorkOrderStatusLogStore {
	return &WorkOrderStatusLogStore{q: q}
}

func (s *WorkOrderStatusLogStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.WorkOrderStatusLogResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListWorkOrderStatusLogs(ctx, gen.ListWorkOrderStatusLogsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountWorkOrderStatusLogs(ctx, gen.CountWorkOrderStatusLogsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.WorkOrderStatusLogResponse, len(rows))
	for i, r := range rows {
		out[i] = toWorkOrderStatusLogResponse(r)
	}
	return out, total, nil
}

func (s *WorkOrderStatusLogStore) Get(ctx context.Context, parentID, id int64) (dto.WorkOrderStatusLogResponse, error) {
	r, err := s.q.GetWorkOrderStatusLog(ctx, gen.GetWorkOrderStatusLogParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.WorkOrderStatusLogResponse{}, err
	}
	return toWorkOrderStatusLogResponse(r), nil
}

func toWorkOrderStatusLogResponse(r gen.WorkOrderStatusLog) dto.WorkOrderStatusLogResponse {
	return dto.WorkOrderStatusLogResponse{
		ID:              r.ID,
		WorkOrderID:     r.WorkOrderID,
		StatusID:        r.StatusID,
		ChangedAt:       r.ChangedAt,
		ActorEmployeeID: r.ActorEmployeeID,
		ActorType:       r.ActorType,
	}
}
