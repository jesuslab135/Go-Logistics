package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type LaborTimeEntryStore struct{ q *gen.Queries }

func NewLaborTimeEntryStore(q *gen.Queries) *LaborTimeEntryStore { return &LaborTimeEntryStore{q: q} }

func (s *LaborTimeEntryStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.LaborTimeEntryResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListLaborTimeEntries(ctx, gen.ListLaborTimeEntriesParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountLaborTimeEntries(ctx, gen.CountLaborTimeEntriesParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.LaborTimeEntryResponse, len(rows))
	for i, r := range rows {
		out[i] = toLaborTimeEntryResponse(r)
	}
	return out, total, nil
}

func (s *LaborTimeEntryStore) Get(ctx context.Context, parentID, id int64) (dto.LaborTimeEntryResponse, error) {
	r, err := s.q.GetLaborTimeEntry(ctx, gen.GetLaborTimeEntryParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.LaborTimeEntryResponse{}, err
	}
	return toLaborTimeEntryResponse(r), nil
}

func (s *LaborTimeEntryStore) Create(ctx context.Context, parentID int64, in dto.CreateLaborTimeEntryRequest) (dto.LaborTimeEntryResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateLaborTimeEntry(ctx, gen.CreateLaborTimeEntryParams{
		ParentID:          parentID,
		CompanyID:         middleware.CompanyFromContext(ctx),
		TechnicianID:      in.TechnicianID,
		StartedAt:         in.StartedAt,
		EndedAt:           in.EndedAt,
		DurationSeconds:   in.DurationSeconds,
		IsActive:          boolOrDefault(in.IsActive, true),
		ClockInLatitude:   in.ClockInLatitude,
		ClockInLongitude:  in.ClockInLongitude,
		ClockOutLatitude:  in.ClockOutLatitude,
		ClockOutLongitude: in.ClockOutLongitude,
		CreatedAt:         now,
	})
	if err != nil {
		return dto.LaborTimeEntryResponse{}, err
	}
	return toLaborTimeEntryResponse(r), nil
}

func (s *LaborTimeEntryStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateLaborTimeEntryRequest) (dto.LaborTimeEntryResponse, error) {
	r, err := s.q.UpdateLaborTimeEntry(ctx, gen.UpdateLaborTimeEntryParams{
		ID:                id,
		ParentID:          parentID,
		CompanyID:         middleware.CompanyFromContext(ctx),
		TechnicianID:      in.TechnicianID,
		StartedAt:         in.StartedAt,
		EndedAt:           in.EndedAt,
		DurationSeconds:   in.DurationSeconds,
		IsActive:          in.IsActive,
		ClockInLatitude:   in.ClockInLatitude,
		ClockInLongitude:  in.ClockInLongitude,
		ClockOutLatitude:  in.ClockOutLatitude,
		ClockOutLongitude: in.ClockOutLongitude,
	})
	if err != nil {
		return dto.LaborTimeEntryResponse{}, err
	}
	return toLaborTimeEntryResponse(r), nil
}

func (s *LaborTimeEntryStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteLaborTimeEntry(ctx, gen.DeleteLaborTimeEntryParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toLaborTimeEntryResponse(r gen.LaborTimeEntry) dto.LaborTimeEntryResponse {
	return dto.LaborTimeEntryResponse{
		ID:                r.ID,
		SubLineItemID:     r.SubLineItemID,
		TechnicianID:      r.TechnicianID,
		StartedAt:         r.StartedAt,
		EndedAt:           r.EndedAt,
		DurationSeconds:   r.DurationSeconds,
		IsActive:          r.IsActive,
		ClockInLatitude:   r.ClockInLatitude,
		ClockInLongitude:  r.ClockInLongitude,
		ClockOutLatitude:  r.ClockOutLatitude,
		ClockOutLongitude: r.ClockOutLongitude,
		CreatedAt:         r.CreatedAt,
	}
}
