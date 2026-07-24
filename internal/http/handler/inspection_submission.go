package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type InspectionSubmissionStore struct{ q *gen.Queries }

func NewInspectionSubmissionStore(q *gen.Queries) *InspectionSubmissionStore {
	return &InspectionSubmissionStore{q: q}
}

func (s *InspectionSubmissionStore) List(ctx context.Context, p paginate.Params) ([]dto.InspectionSubmissionResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListInspectionSubmissions(ctx, gen.ListInspectionSubmissionsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountInspectionSubmissions(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.InspectionSubmissionResponse, len(rows))
	for i, r := range rows {
		out[i] = toInspectionSubmissionResponse(r)
	}
	return out, total, nil
}

func (s *InspectionSubmissionStore) Get(ctx context.Context, id int64) (dto.InspectionSubmissionResponse, error) {
	r, err := s.q.GetInspectionSubmission(ctx, gen.GetInspectionSubmissionParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.InspectionSubmissionResponse{}, err
	}
	return toInspectionSubmissionResponse(r), nil
}

func (s *InspectionSubmissionStore) Create(ctx context.Context, in dto.CreateInspectionSubmissionRequest) (dto.InspectionSubmissionResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateInspectionSubmission(ctx, gen.CreateInspectionSubmissionParams{
		CompanyID:          middleware.CompanyFromContext(ctx),
		FormID:             in.FormID,
		AssetID:            in.AssetID,
		SubmittedByID:      in.SubmittedByID,
		StartedAt:          in.StartedAt,
		SubmittedAt:        in.SubmittedAt,
		DurationSeconds:    in.DurationSeconds,
		StartingLatitude:   in.StartingLatitude,
		StartingLongitude:  in.StartingLongitude,
		SubmittedLatitude:  in.SubmittedLatitude,
		SubmittedLongitude: in.SubmittedLongitude,
		Signature:          in.Signature,
		Odometer:           in.Odometer,
		TotalItems:         in.TotalItems,
		FailedItemsCount:   in.FailedItemsCount,
		PassedItemsCount:   in.PassedItemsCount,
		CommentsCount:      in.CommentsCount,
		ImagesCount:        in.ImagesCount,
		GeneralNotes:       in.GeneralNotes,
		CreatedAt:          now,
	})
	if err != nil {
		return dto.InspectionSubmissionResponse{}, err
	}
	return toInspectionSubmissionResponse(r), nil
}

func (s *InspectionSubmissionStore) Update(ctx context.Context, id int64, in dto.UpdateInspectionSubmissionRequest) (dto.InspectionSubmissionResponse, error) {
	r, err := s.q.UpdateInspectionSubmission(ctx, gen.UpdateInspectionSubmissionParams{
		ID:                 id,
		CompanyID:          middleware.CompanyFromContext(ctx),
		FormID:             in.FormID,
		AssetID:            in.AssetID,
		SubmittedByID:      in.SubmittedByID,
		StartedAt:          in.StartedAt,
		SubmittedAt:        in.SubmittedAt,
		DurationSeconds:    in.DurationSeconds,
		StartingLatitude:   in.StartingLatitude,
		StartingLongitude:  in.StartingLongitude,
		SubmittedLatitude:  in.SubmittedLatitude,
		SubmittedLongitude: in.SubmittedLongitude,
		Signature:          in.Signature,
		Odometer:           in.Odometer,
		TotalItems:         in.TotalItems,
		FailedItemsCount:   in.FailedItemsCount,
		PassedItemsCount:   in.PassedItemsCount,
		CommentsCount:      in.CommentsCount,
		ImagesCount:        in.ImagesCount,
		GeneralNotes:       in.GeneralNotes,
	})
	if err != nil {
		return dto.InspectionSubmissionResponse{}, err
	}
	return toInspectionSubmissionResponse(r), nil
}

func (s *InspectionSubmissionStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteInspectionSubmission(ctx, gen.DeleteInspectionSubmissionParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toInspectionSubmissionResponse(r gen.InspectionSubmission) dto.InspectionSubmissionResponse {
	return dto.InspectionSubmissionResponse{
		ID:                 r.ID,
		CompanyID:          r.CompanyID,
		FormID:             r.FormID,
		AssetID:            r.AssetID,
		SubmittedByID:      r.SubmittedByID,
		StartedAt:          r.StartedAt,
		SubmittedAt:        r.SubmittedAt,
		DurationSeconds:    r.DurationSeconds,
		StartingLatitude:   r.StartingLatitude,
		StartingLongitude:  r.StartingLongitude,
		SubmittedLatitude:  r.SubmittedLatitude,
		SubmittedLongitude: r.SubmittedLongitude,
		Signature:          r.Signature,
		Odometer:           r.Odometer,
		TotalItems:         r.TotalItems,
		FailedItemsCount:   r.FailedItemsCount,
		PassedItemsCount:   r.PassedItemsCount,
		CommentsCount:      r.CommentsCount,
		ImagesCount:        r.ImagesCount,
		GeneralNotes:       r.GeneralNotes,
		CreatedAt:          r.CreatedAt,
	}
}
