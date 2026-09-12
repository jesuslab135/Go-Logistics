package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/domain/tire"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/dbctx"
	"fleet/internal/platform/paginate"
)

type TireAssignmentRequestStore struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewTireAssignmentRequestStore(q *gen.Queries, pool *pgxpool.Pool) *TireAssignmentRequestStore {
	return &TireAssignmentRequestStore{q: q, pool: pool}
}

func (s *TireAssignmentRequestStore) List(ctx context.Context, p paginate.Params) ([]dto.TireAssignmentRequestResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListTireAssignmentRequests(ctx, gen.ListTireAssignmentRequestsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountTireAssignmentRequests(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.TireAssignmentRequestResponse, len(rows))
	for i, r := range rows {
		out[i] = toTireAssignmentRequestResponse(r)
	}
	return out, total, nil
}

func (s *TireAssignmentRequestStore) Get(ctx context.Context, id int64) (dto.TireAssignmentRequestResponse, error) {
	r, err := s.q.GetTireAssignmentRequest(ctx, gen.GetTireAssignmentRequestParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.TireAssignmentRequestResponse{}, err
	}
	return toTireAssignmentRequestResponse(r), nil
}

func (s *TireAssignmentRequestStore) Create(ctx context.Context, in dto.CreateTireAssignmentRequestRequest) (dto.TireAssignmentRequestResponse, error) {
	company := middleware.CompanyFromContext(ctx)
	now := time.Now().UTC()

	tx, err := dbctx.Begin(ctx, s.pool)
	if err != nil {
		return dto.TireAssignmentRequestResponse{}, err
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	r, err := qtx.CreateTireAssignmentRequest(ctx, gen.CreateTireAssignmentRequestParams{
		CompanyID:     company,
		TireID:        in.TireID,
		VehicleID:     in.VehicleID,
		PositionCode:  in.PositionCode,
		State:         tire.RequestPending,
		RequestedByID: authorFromContext(ctx),
		RequestedAt:   now,
		Notes:         in.Notes,
	})
	if err != nil {
		return dto.TireAssignmentRequestResponse{}, err
	}

	// A request sitting unseen in an inbox is the reason this endpoint exists,
	// so everyone who can approve it is notified in the same transaction.
	if err := qtx.NotifyTireApprovers(ctx, gen.NotifyTireApproversParams{
		CompanyID: company,
		Kind:      NotificationTireAssignmentPending,
		Title:     "Tire assignment request",
		Body:      fmt.Sprintf("Tire %d requested for vehicle %d at position %s.", in.TireID, in.VehicleID, in.PositionCode),
		Url:       fmt.Sprintf("/tire-assignment-requests/%d", r.ID),
		CreatedAt: now,
	}); err != nil {
		return dto.TireAssignmentRequestResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto.TireAssignmentRequestResponse{}, err
	}
	return toTireAssignmentRequestResponse(r), nil
}

func (s *TireAssignmentRequestStore) Update(ctx context.Context, id int64, in dto.UpdateTireAssignmentRequestRequest) (dto.TireAssignmentRequestResponse, error) {
	r, err := s.q.UpdateTireAssignmentRequest(ctx, gen.UpdateTireAssignmentRequestParams{
		ID:              id,
		CompanyID:       middleware.CompanyFromContext(ctx),
		TireID:          in.TireID,
		VehicleID:       in.VehicleID,
		PositionCode:    in.PositionCode,
		State:           in.State,
		RequestedByID:   in.RequestedByID,
		RequestedAt:     in.RequestedAt,
		ApprovedByID:    in.ApprovedByID,
		ResolvedAt:      in.ResolvedAt,
		RejectionReason: in.RejectionReason,
		Notes:           in.Notes,
	})
	if err != nil {
		return dto.TireAssignmentRequestResponse{}, err
	}
	return toTireAssignmentRequestResponse(r), nil
}

func (s *TireAssignmentRequestStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteTireAssignmentRequest(ctx, gen.DeleteTireAssignmentRequestParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toTireAssignmentRequestResponse(r gen.TireAssignmentRequest) dto.TireAssignmentRequestResponse {
	return dto.TireAssignmentRequestResponse{
		ID:              r.ID,
		CompanyID:       r.CompanyID,
		TireID:          r.TireID,
		VehicleID:       r.VehicleID,
		PositionCode:    r.PositionCode,
		State:           r.State,
		RequestedByID:   r.RequestedByID,
		RequestedAt:     r.RequestedAt,
		ApprovedByID:    r.ApprovedByID,
		ResolvedAt:      r.ResolvedAt,
		RejectionReason: r.RejectionReason,
		Notes:           r.Notes,
	}
}
