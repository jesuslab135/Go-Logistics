package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type InventoryAdjustmentReasonStore struct{ q *gen.Queries }

func NewInventoryAdjustmentReasonStore(q *gen.Queries) *InventoryAdjustmentReasonStore {
	return &InventoryAdjustmentReasonStore{q: q}
}

func (s *InventoryAdjustmentReasonStore) List(ctx context.Context, p paginate.Params) ([]dto.InventoryAdjustmentReasonResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListInventoryAdjustmentReasons(ctx, gen.ListInventoryAdjustmentReasonsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountInventoryAdjustmentReasons(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.InventoryAdjustmentReasonResponse, len(rows))
	for i, r := range rows {
		out[i] = toInventoryAdjustmentReasonResponse(r)
	}
	return out, total, nil
}

func (s *InventoryAdjustmentReasonStore) Get(ctx context.Context, id int64) (dto.InventoryAdjustmentReasonResponse, error) {
	r, err := s.q.GetInventoryAdjustmentReason(ctx, gen.GetInventoryAdjustmentReasonParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.InventoryAdjustmentReasonResponse{}, err
	}
	return toInventoryAdjustmentReasonResponse(r), nil
}

func (s *InventoryAdjustmentReasonStore) Create(ctx context.Context, in dto.CreateInventoryAdjustmentReasonRequest) (dto.InventoryAdjustmentReasonResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateInventoryAdjustmentReason(ctx, gen.CreateInventoryAdjustmentReasonParams{
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
		CreatedAt: now,
	})
	if err != nil {
		return dto.InventoryAdjustmentReasonResponse{}, err
	}
	return toInventoryAdjustmentReasonResponse(r), nil
}

func (s *InventoryAdjustmentReasonStore) Update(ctx context.Context, id int64, in dto.UpdateInventoryAdjustmentReasonRequest) (dto.InventoryAdjustmentReasonResponse, error) {
	r, err := s.q.UpdateInventoryAdjustmentReason(ctx, gen.UpdateInventoryAdjustmentReasonParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
		Name:      in.Name,
	})
	if err != nil {
		return dto.InventoryAdjustmentReasonResponse{}, err
	}
	return toInventoryAdjustmentReasonResponse(r), nil
}

func (s *InventoryAdjustmentReasonStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteInventoryAdjustmentReason(ctx, gen.DeleteInventoryAdjustmentReasonParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toInventoryAdjustmentReasonResponse(r gen.InventoryAdjustmentReason) dto.InventoryAdjustmentReasonResponse {
	return dto.InventoryAdjustmentReasonResponse{
		ID:        r.ID,
		CompanyID: r.CompanyID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt,
	}
}
