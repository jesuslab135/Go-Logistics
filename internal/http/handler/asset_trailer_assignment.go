package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type AssetTrailerAssignmentStore struct{ q *gen.Queries }

func NewAssetTrailerAssignmentStore(q *gen.Queries) *AssetTrailerAssignmentStore {
	return &AssetTrailerAssignmentStore{q: q}
}

func (s *AssetTrailerAssignmentStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.AssetTrailerAssignmentResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListAssetTrailerAssignments(ctx, gen.ListAssetTrailerAssignmentsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountAssetTrailerAssignments(ctx, gen.CountAssetTrailerAssignmentsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.AssetTrailerAssignmentResponse, len(rows))
	for i, r := range rows {
		out[i] = toAssetTrailerAssignmentResponse(r)
	}
	return out, total, nil
}

func (s *AssetTrailerAssignmentStore) Get(ctx context.Context, parentID, id int64) (dto.AssetTrailerAssignmentResponse, error) {
	r, err := s.q.GetAssetTrailerAssignment(ctx, gen.GetAssetTrailerAssignmentParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.AssetTrailerAssignmentResponse{}, err
	}
	return toAssetTrailerAssignmentResponse(r), nil
}

func (s *AssetTrailerAssignmentStore) Create(ctx context.Context, parentID int64, in dto.CreateAssetTrailerAssignmentRequest) (dto.AssetTrailerAssignmentResponse, error) {
	r, err := s.q.CreateAssetTrailerAssignment(ctx, gen.CreateAssetTrailerAssignmentParams{
		ParentID:       parentID,
		CompanyID:      middleware.CompanyFromContext(ctx),
		TrailerID:      in.TrailerID,
		Position:       int32OrDefault(in.Position, 1),
		AssignedDate:   in.AssignedDate,
		UnassignedDate: in.UnassignedDate,
		AssignedByID:   in.AssignedByID,
		IsActive:       boolOrDefault(in.IsActive, true),
		Notes:          in.Notes,
	})
	if err != nil {
		return dto.AssetTrailerAssignmentResponse{}, err
	}
	return toAssetTrailerAssignmentResponse(r), nil
}

func (s *AssetTrailerAssignmentStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateAssetTrailerAssignmentRequest) (dto.AssetTrailerAssignmentResponse, error) {
	r, err := s.q.UpdateAssetTrailerAssignment(ctx, gen.UpdateAssetTrailerAssignmentParams{
		ID:             id,
		ParentID:       parentID,
		CompanyID:      middleware.CompanyFromContext(ctx),
		TrailerID:      in.TrailerID,
		Position:       in.Position,
		AssignedDate:   in.AssignedDate,
		UnassignedDate: in.UnassignedDate,
		AssignedByID:   in.AssignedByID,
		IsActive:       in.IsActive,
		Notes:          in.Notes,
	})
	if err != nil {
		return dto.AssetTrailerAssignmentResponse{}, err
	}
	return toAssetTrailerAssignmentResponse(r), nil
}

func (s *AssetTrailerAssignmentStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteAssetTrailerAssignment(ctx, gen.DeleteAssetTrailerAssignmentParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toAssetTrailerAssignmentResponse(r gen.AssetTrailerAssignment) dto.AssetTrailerAssignmentResponse {
	return dto.AssetTrailerAssignmentResponse{
		ID:             r.ID,
		AssetID:        r.AssetID,
		TrailerID:      r.TrailerID,
		Position:       r.Position,
		AssignedDate:   r.AssignedDate,
		UnassignedDate: r.UnassignedDate,
		AssignedByID:   r.AssignedByID,
		IsActive:       r.IsActive,
		Notes:          r.Notes,
	}
}
