package handler

import (
	"context"
	"encoding/json"
	"log/slog"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
	"fleet/internal/platform/storage"
)

type InspectionSubmissionItemStore struct {
	fileOwner
	q *gen.Queries
}

func NewInspectionSubmissionItemStore(q *gen.Queries, files storage.Storage, log *slog.Logger) *InspectionSubmissionItemStore {
	return &InspectionSubmissionItemStore{fileOwner: newFileOwner(files, log), q: q}
}

func (s *InspectionSubmissionItemStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.InspectionSubmissionItemResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListInspectionSubmissionItems(ctx, gen.ListInspectionSubmissionItemsParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountInspectionSubmissionItems(ctx, gen.CountInspectionSubmissionItemsParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.InspectionSubmissionItemResponse, len(rows))
	for i, r := range rows {
		out[i] = toInspectionSubmissionItemResponse(r)
	}
	return out, total, nil
}

func (s *InspectionSubmissionItemStore) Get(ctx context.Context, parentID, id int64) (dto.InspectionSubmissionItemResponse, error) {
	r, err := s.q.GetInspectionSubmissionItem(ctx, gen.GetInspectionSubmissionItemParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.InspectionSubmissionItemResponse{}, err
	}
	return toInspectionSubmissionItemResponse(r), nil
}

func (s *InspectionSubmissionItemStore) Create(ctx context.Context, parentID int64, in dto.CreateInspectionSubmissionItemRequest) (dto.InspectionSubmissionItemResponse, error) {
	r, err := s.q.CreateInspectionSubmissionItem(ctx, gen.CreateInspectionSubmissionItemParams{
		ParentID:         parentID,
		CompanyID:        middleware.CompanyFromContext(ctx),
		FormItemID:       in.FormItemID,
		ResultStatus:     in.ResultStatus,
		ResultValue:      jsonbOrDefault(in.ResultValue, "{}"),
		Remark:           in.Remark,
		Photo:            in.Photo,
		Latitude:         in.Latitude,
		Longitude:        in.Longitude,
		GeneratedIssueID: in.GeneratedIssueID,
	})
	if err != nil {
		return dto.InspectionSubmissionItemResponse{}, err
	}
	return toInspectionSubmissionItemResponse(r), nil
}

func (s *InspectionSubmissionItemStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateInspectionSubmissionItemRequest) (dto.InspectionSubmissionItemResponse, error) {
	previous, err := s.q.GetInspectionSubmissionItem(ctx, gen.GetInspectionSubmissionItemParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.InspectionSubmissionItemResponse{}, err
	}

	r, err := s.q.UpdateInspectionSubmissionItem(ctx, gen.UpdateInspectionSubmissionItemParams{
		ID:               id,
		ParentID:         parentID,
		CompanyID:        middleware.CompanyFromContext(ctx),
		FormItemID:       in.FormItemID,
		ResultStatus:     in.ResultStatus,
		ResultValue:      jsonbOrDefault(in.ResultValue, "{}"),
		Remark:           in.Remark,
		Photo:            in.Photo,
		Latitude:         in.Latitude,
		Longitude:        in.Longitude,
		GeneratedIssueID: in.GeneratedIssueID,
	})
	if err != nil {
		return dto.InspectionSubmissionItemResponse{}, err
	}
	s.reclaimReplaced(ctx, previous.Photo, r.Photo)
	return toInspectionSubmissionItemResponse(r), nil
}

func (s *InspectionSubmissionItemStore) Delete(ctx context.Context, parentID, id int64) error {
	company := middleware.CompanyFromContext(ctx)

	previous, err := s.q.GetInspectionSubmissionItem(ctx, gen.GetInspectionSubmissionItemParams{ID: id, ParentID: parentID, CompanyID: company})
	if err != nil {
		return err
	}
	if err := s.q.DeleteInspectionSubmissionItem(ctx, gen.DeleteInspectionSubmissionItemParams{ID: id, ParentID: parentID, CompanyID: company}); err != nil {
		return err
	}
	if previous.Photo != nil {
		s.reclaim(ctx, *previous.Photo)
	}
	return nil
}

func toInspectionSubmissionItemResponse(r gen.InspectionSubmissionItem) dto.InspectionSubmissionItemResponse {
	return dto.InspectionSubmissionItemResponse{
		ID:               r.ID,
		SubmissionID:     r.SubmissionID,
		FormItemID:       r.FormItemID,
		ResultStatus:     r.ResultStatus,
		ResultValue:      json.RawMessage(r.ResultValue),
		Remark:           r.Remark,
		Photo:            r.Photo,
		Latitude:         r.Latitude,
		Longitude:        r.Longitude,
		GeneratedIssueID: r.GeneratedIssueID,
	}
}
