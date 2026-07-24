package handler

import (
	"context"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type MediumStore struct{ q *gen.Queries }

func NewMediumStore(q *gen.Queries) *MediumStore { return &MediumStore{q: q} }

func (s *MediumStore) List(ctx context.Context, p paginate.Params) ([]dto.MediumResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListMediaItems(ctx, gen.ListMediaItemsParams{CompanyID: company, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountMediaItems(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.MediumResponse, len(rows))
	for i, r := range rows {
		out[i] = toMediumResponse(r)
	}
	return out, total, nil
}

func (s *MediumStore) Get(ctx context.Context, id int64) (dto.MediumResponse, error) {
	r, err := s.q.GetMedium(ctx, gen.GetMediumParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.MediumResponse{}, err
	}
	return toMediumResponse(r), nil
}

func (s *MediumStore) Create(ctx context.Context, in dto.CreateMediumRequest) (dto.MediumResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.CreateMedium(ctx, gen.CreateMediumParams{
		CompanyID:    middleware.CompanyFromContext(ctx),
		AssetID:      in.AssetID,
		File:         in.File,
		Title:        in.Title,
		Description:  in.Description,
		FileType:     in.FileType,
		FileSize:     in.FileSize,
		UploadedByID: in.UploadedByID,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return dto.MediumResponse{}, err
	}
	return toMediumResponse(r), nil
}

func (s *MediumStore) Update(ctx context.Context, id int64, in dto.UpdateMediumRequest) (dto.MediumResponse, error) {
	now := time.Now().UTC()
	r, err := s.q.UpdateMedium(ctx, gen.UpdateMediumParams{
		ID:           id,
		CompanyID:    middleware.CompanyFromContext(ctx),
		AssetID:      in.AssetID,
		File:         in.File,
		Title:        in.Title,
		Description:  in.Description,
		FileType:     in.FileType,
		FileSize:     in.FileSize,
		UploadedByID: in.UploadedByID,
		UpdatedAt:    now,
	})
	if err != nil {
		return dto.MediumResponse{}, err
	}
	return toMediumResponse(r), nil
}

func (s *MediumStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteMedium(ctx, gen.DeleteMediumParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toMediumResponse(r gen.Medium) dto.MediumResponse {
	return dto.MediumResponse{
		ID:           r.ID,
		CompanyID:    r.CompanyID,
		AssetID:      r.AssetID,
		File:         r.File,
		Title:        r.Title,
		Description:  r.Description,
		FileType:     r.FileType,
		FileSize:     r.FileSize,
		UploadedByID: r.UploadedByID,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}
