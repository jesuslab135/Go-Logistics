package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type FuelCommentStore struct{ q *gen.Queries }

func NewFuelCommentStore(q *gen.Queries) *FuelCommentStore { return &FuelCommentStore{q: q} }

func (s *FuelCommentStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.FuelCommentResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListFuelComments(ctx, gen.ListFuelCommentsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountFuelComments(ctx, gen.CountFuelCommentsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.FuelCommentResponse, len(rows))
	for i, r := range rows {
		out[i] = toFuelCommentResponse(r)
	}
	return out, total, nil
}

func (s *FuelCommentStore) Get(ctx context.Context, parentID, id int64) (dto.FuelCommentResponse, error) {
	r, err := s.q.GetFuelComment(ctx, gen.GetFuelCommentParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.FuelCommentResponse{}, err
	}
	return toFuelCommentResponse(r), nil
}

func (s *FuelCommentStore) Create(ctx context.Context, parentID int64, in dto.CreateFuelCommentRequest) (dto.FuelCommentResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateFuelComment(ctx, gen.CreateFuelCommentParams{
		ParentID:   parentID,
		CompanyID:  middleware.CompanyFromContext(ctx),
		EmployeeID: authorFromContext(ctx),
		Text:       in.Text,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if err != nil {
		return dto.FuelCommentResponse{}, err
	}
	return toFuelCommentResponse(r), nil
}

func (s *FuelCommentStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateFuelCommentRequest) (dto.FuelCommentResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateFuelComment(ctx, gen.UpdateFuelCommentParams{
		ID:        id,
		ParentID:  parentID,
		CompanyID: middleware.CompanyFromContext(ctx),
		Text:      in.Text,
		UpdatedAt: now,
	})
	if err != nil {
		return dto.FuelCommentResponse{}, err
	}
	return toFuelCommentResponse(r), nil
}

func (s *FuelCommentStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteFuelComment(ctx, gen.DeleteFuelCommentParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toFuelCommentResponse(r gen.FuelComment) dto.FuelCommentResponse {
	return dto.FuelCommentResponse{
		ID:         r.ID,
		EntryID:    r.EntryID,
		EmployeeID: r.EmployeeID,
		Text:       r.Text,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}
