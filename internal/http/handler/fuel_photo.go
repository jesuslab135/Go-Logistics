package handler

import (
	"context"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
)

type FuelPhotoStore struct{ q *gen.Queries }

func NewFuelPhotoStore(q *gen.Queries) *FuelPhotoStore { return &FuelPhotoStore{q: q} }

func (s *FuelPhotoStore) List(ctx context.Context, parentID int64, p paginate.Params) ([]dto.FuelPhotoResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListFuelPhotos(ctx, gen.ListFuelPhotosParams{ParentID: parentID, CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountFuelPhotos(ctx, gen.CountFuelPhotosParams{ParentID: parentID, CompanyID: company})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.FuelPhotoResponse, len(rows))
	for i, r := range rows {
		out[i] = toFuelPhotoResponse(r)
	}
	return out, total, nil
}

func (s *FuelPhotoStore) Get(ctx context.Context, parentID, id int64) (dto.FuelPhotoResponse, error) {
	r, err := s.q.GetFuelPhoto(ctx, gen.GetFuelPhotoParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.FuelPhotoResponse{}, err
	}
	return toFuelPhotoResponse(r), nil
}

func (s *FuelPhotoStore) Create(ctx context.Context, parentID int64, in dto.CreateFuelPhotoRequest) (dto.FuelPhotoResponse, error) {
	r, err := s.q.CreateFuelPhoto(ctx, gen.CreateFuelPhotoParams{
		ParentID:     parentID,
		CompanyID:    middleware.CompanyFromContext(ctx),
		UploadedByID: authorFromContext(ctx),
		File:         in.File,
		FileName:     in.FileName,
		FileSize:     in.FileSize,
		MimeType:     in.MimeType,
		Description:  in.Description,
		UploadedAt:   in.UploadedAt,
		IsPrimary:    in.IsPrimary,
	})
	if err != nil {
		return dto.FuelPhotoResponse{}, err
	}
	return toFuelPhotoResponse(r), nil
}

func (s *FuelPhotoStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateFuelPhotoRequest) (dto.FuelPhotoResponse, error) {
	r, err := s.q.UpdateFuelPhoto(ctx, gen.UpdateFuelPhotoParams{
		ID:          id,
		ParentID:    parentID,
		CompanyID:   middleware.CompanyFromContext(ctx),
		File:        in.File,
		FileName:    in.FileName,
		FileSize:    in.FileSize,
		MimeType:    in.MimeType,
		Description: in.Description,
		UploadedAt:  in.UploadedAt,
		IsPrimary:   in.IsPrimary,
	})
	if err != nil {
		return dto.FuelPhotoResponse{}, err
	}
	return toFuelPhotoResponse(r), nil
}

func (s *FuelPhotoStore) Delete(ctx context.Context, parentID, id int64) error {
	return s.q.DeleteFuelPhoto(ctx, gen.DeleteFuelPhotoParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
}

func toFuelPhotoResponse(r gen.FuelPhoto) dto.FuelPhotoResponse {
	return dto.FuelPhotoResponse{
		ID:           r.ID,
		EntryID:      r.EntryID,
		UploadedByID: r.UploadedByID,
		File:         r.File,
		FileName:     r.FileName,
		FileSize:     r.FileSize,
		MimeType:     r.MimeType,
		Description:  r.Description,
		UploadedAt:   r.UploadedAt,
		IsPrimary:    r.IsPrimary,
	}
}
