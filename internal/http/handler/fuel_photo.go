package handler

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/paginate"
	"fleet/internal/platform/storage"
)

type FuelPhotoStore struct {
	fileOwner
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewFuelPhotoStore(q *gen.Queries, pool *pgxpool.Pool, files storage.Storage, log *slog.Logger) *FuelPhotoStore {
	return &FuelPhotoStore{fileOwner: newFileOwner(files, log), q: q, pool: pool}
}

// SetPrimary promotes one photo and demotes the rest, in one transaction.
//
// At most one photo per entry may be primary, and a partial unique index
// enforces it — so promoting without demoting first is not a race that produces
// two primaries, it is a constraint violation. Doing both in one transaction is
// what makes the operation possible at all, which is why is_primary is not a
// field on create or update.
//
// Deleting the primary photo promotes nothing. Auto-promotion would silently
// designate a photo nobody chose, and "no primary" is a state every consumer
// already has to render, because an entry with no photos at all is valid.
func (s *FuelPhotoStore) SetPrimary(ctx context.Context, parentID, id int64) (dto.FuelPhotoResponse, error) {
	company := middleware.CompanyFromContext(ctx)

	var out dto.FuelPhotoResponse
	err := inTx(ctx, s.pool, s.q, func(qtx *gen.Queries) error {
		if err := qtx.ClearFuelEntryPrimaryPhoto(ctx, gen.ClearFuelEntryPrimaryPhotoParams{
			ParentID: parentID, CompanyID: company,
		}); err != nil {
			return err
		}
		r, err := qtx.SetFuelPhotoPrimary(ctx, gen.SetFuelPhotoPrimaryParams{
			ID: id, ParentID: parentID, CompanyID: company,
		})
		if err != nil {
			return err
		}
		out = s.response(ctx, r)
		return nil
	})
	return out, err
}

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
		out[i] = s.response(ctx, r)
	}
	return out, total, nil
}

func (s *FuelPhotoStore) Get(ctx context.Context, parentID, id int64) (dto.FuelPhotoResponse, error) {
	r, err := s.q.GetFuelPhoto(ctx, gen.GetFuelPhotoParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.FuelPhotoResponse{}, err
	}
	return s.response(ctx, r), nil
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
	})
	if err != nil {
		return dto.FuelPhotoResponse{}, err
	}
	return s.response(ctx, r), nil
}

func (s *FuelPhotoStore) Update(ctx context.Context, parentID, id int64, in dto.UpdateFuelPhotoRequest) (dto.FuelPhotoResponse, error) {
	previous, err := s.q.GetFuelPhoto(ctx, gen.GetFuelPhotoParams{ID: id, ParentID: parentID, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.FuelPhotoResponse{}, err
	}

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
	})
	if err != nil {
		return dto.FuelPhotoResponse{}, err
	}
	s.reclaimReplaced(ctx, &previous.File, &r.File)
	return s.response(ctx, r), nil
}

func (s *FuelPhotoStore) Delete(ctx context.Context, parentID, id int64) error {
	company := middleware.CompanyFromContext(ctx)

	previous, err := s.q.GetFuelPhoto(ctx, gen.GetFuelPhotoParams{ID: id, ParentID: parentID, CompanyID: company})
	if err != nil {
		return err
	}
	if err := s.q.DeleteFuelPhoto(ctx, gen.DeleteFuelPhotoParams{ID: id, ParentID: parentID, CompanyID: company}); err != nil {
		return err
	}
	s.reclaim(ctx, previous.File)
	return nil
}

// response resolves the stored object reference to a URL the caller can read.
// A fuel receipt is private, so that URL is signed and expires; it is produced
// here, per response, and never persisted.
func (s *FuelPhotoStore) response(ctx context.Context, r gen.FuelPhoto) dto.FuelPhotoResponse {
	out := toFuelPhotoResponse(r)
	out.File, out.FileExpiresAt = fileReadURL(ctx, s.files, r.File)
	return out
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
