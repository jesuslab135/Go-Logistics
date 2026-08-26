package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

// PurchaseOrderStatusLogStore reads an order's transition history. As with the
// work-order status log, it has no write methods: rows are appended by the
// transition that caused them, inside that transaction, so the history cannot
// disagree with the order it describes.
type PurchaseOrderStatusLogStore struct{ q *gen.Queries }

func NewPurchaseOrderStatusLogStore(q *gen.Queries) *PurchaseOrderStatusLogStore {
	return &PurchaseOrderStatusLogStore{q: q}
}

func (s *PurchaseOrderStatusLogStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.PurchaseOrderStatusLogResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListPurchaseOrderStatusLogs(ctx, gen.ListPurchaseOrderStatusLogsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountPurchaseOrderStatusLogs(ctx, gen.CountPurchaseOrderStatusLogsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PurchaseOrderStatusLogResponse, len(rows))
	for i, r := range rows {
		out[i] = toPurchaseOrderStatusLogResponse(r)
	}
	return out, total, nil
}

func (s *PurchaseOrderStatusLogStore) Get(ctx context.Context, parentID, id int64) (dto.PurchaseOrderStatusLogResponse, error) {
	r, err := s.q.GetPurchaseOrderStatusLog(ctx, gen.GetPurchaseOrderStatusLogParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.PurchaseOrderStatusLogResponse{}, err
	}
	return toPurchaseOrderStatusLogResponse(r), nil
}

func toPurchaseOrderStatusLogResponse(r gen.PurchaseOrderStatusLog) dto.PurchaseOrderStatusLogResponse {
	return dto.PurchaseOrderStatusLogResponse{
		ID:              r.ID,
		PurchaseOrderID: r.PurchaseOrderID,
		FromState:       r.FromState,
		ToState:         r.ToState,
		ActorEmployeeID: r.ActorEmployeeID,
		ActorType:       r.ActorType,
		Reason:          r.Reason,
		ChangedAt:       r.ChangedAt,
	}
}
