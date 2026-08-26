package handler

import (
	"context"
	"errors"
	"log/slog"
	"path"
	"strings"
	"time"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/imaging"
	"fleet/internal/platform/paginate"
	"fleet/internal/platform/storage"
)

type MediumStore struct {
	fileOwner
	q     *gen.Queries
	files storage.Storage
}

func NewMediumStore(q *gen.Queries, files storage.Storage, log *slog.Logger) *MediumStore {
	return &MediumStore{fileOwner: newFileOwner(files, log), q: q, files: files}
}

// resolvedFile is the server's own view of a client-supplied file URL.
type resolvedFile struct {
	fileType  string
	size      int32
	thumbnail string
}

// resolveFile mirrors Django's Media.save(): file_type and file_size are derived
// from the stored file, never taken from the request. Resolving the URL back to
// a key also proves it points at an object in our own bucket.
//
// The thumbnail follows the same rule. Rather than assume the derived key
// exists, it is confirmed against storage, so media predating thumbnailing (and
// formats the server cannot decode) report no thumbnail instead of a URL that
// would 404 in the gallery.
func (s *MediumStore) resolveFile(ctx context.Context, url string) (resolvedFile, error) {
	key, ok := s.files.KeyFromURL(url)
	if !ok {
		return resolvedFile{}, apierr.BadRequest("file must be a URL returned by POST /api/v1/uploads")
	}
	obj, err := s.files.Stat(ctx, key)
	if errors.Is(err, storage.ErrNotFound) {
		return resolvedFile{}, apierr.BadRequest("file does not exist in storage")
	}
	if err != nil {
		return resolvedFile{}, err
	}

	out := resolvedFile{fileType: detectFileType(key), size: int32(obj.Size)}
	if out.fileType == "IMAGE" {
		thumbKey := imaging.ThumbKey(key)
		if _, err := s.files.Stat(ctx, thumbKey); err == nil {
			out.thumbnail = s.files.URL(thumbKey)
		}
	}
	return out, nil
}

// detectFileType ports Django's api/models/media_model.py detect_file_type.
func detectFileType(key string) string {
	switch strings.ToLower(path.Ext(key)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return "IMAGE"
	case ".pdf":
		return "PDF"
	case ".doc", ".docx", ".xls", ".xlsx", ".csv", ".txt":
		return "DOCUMENT"
	default:
		return "OTHER"
	}
}

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
		out[i] = s.response(ctx, r)
	}
	return out, total, nil
}

func (s *MediumStore) Get(ctx context.Context, id int64) (dto.MediumResponse, error) {
	r, err := s.q.GetMedium(ctx, gen.GetMediumParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.MediumResponse{}, err
	}
	return s.response(ctx, r), nil
}

func (s *MediumStore) Create(ctx context.Context, in dto.CreateMediumRequest) (dto.MediumResponse, error) {
	file, err := s.resolveFile(ctx, in.File)
	if err != nil {
		return dto.MediumResponse{}, err
	}

	now := time.Now().UTC()
	r, err := s.q.CreateMedium(ctx, gen.CreateMediumParams{
		CompanyID:    middleware.CompanyFromContext(ctx),
		AssetID:      in.AssetID,
		File:         in.File,
		Title:        in.Title,
		Description:  in.Description,
		FileType:     file.fileType,
		FileSize:     file.size,
		Thumbnail:    file.thumbnail,
		UploadedByID: authorFromContext(ctx),
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return dto.MediumResponse{}, err
	}
	return s.response(ctx, r), nil
}

func (s *MediumStore) Update(ctx context.Context, id int64, in dto.UpdateMediumRequest) (dto.MediumResponse, error) {
	now := time.Now().UTC()
	file, err := s.resolveFile(ctx, in.File)
	if err != nil {
		return dto.MediumResponse{}, err
	}

	previous, err := s.q.GetMedium(ctx, gen.GetMediumParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.MediumResponse{}, err
	}

	r, err := s.q.UpdateMedium(ctx, gen.UpdateMediumParams{
		ID:          id,
		CompanyID:   middleware.CompanyFromContext(ctx),
		AssetID:     in.AssetID,
		File:        in.File,
		Title:       in.Title,
		Description: in.Description,
		FileType:    file.fileType,
		FileSize:    file.size,
		Thumbnail:   file.thumbnail,
		UpdatedAt:   now,
	})
	if err != nil {
		return dto.MediumResponse{}, err
	}
	// The row now points elsewhere, so the file it used to reference — and its
	// derived thumbnail — are unreachable.
	if previous.File != r.File {
		s.reclaim(ctx, previous.File, previous.Thumbnail)
	}
	return s.response(ctx, r), nil
}

func (s *MediumStore) Delete(ctx context.Context, id int64) error {
	company := middleware.CompanyFromContext(ctx)

	previous, err := s.q.GetMedium(ctx, gen.GetMediumParams{ID: id, CompanyID: company})
	if err != nil {
		return err
	}
	if err := s.q.DeleteMedium(ctx, gen.DeleteMediumParams{ID: id, CompanyID: company}); err != nil {
		return err
	}
	s.reclaim(ctx, previous.File, previous.Thumbnail)
	return nil
}

// response resolves the stored object references to URLs the caller can read.
// Uploaded media belongs to one tenant, so those URLs are signed and expire;
// they are produced here, per response, and never persisted. The thumbnail
// shares the original's expiry because both are signed in the same call.
func (s *MediumStore) response(ctx context.Context, r gen.Medium) dto.MediumResponse {
	out := toMediumResponse(r)
	out.File, out.FileExpiresAt = fileReadURL(ctx, s.files, r.File)
	if r.Thumbnail != "" {
		out.Thumbnail, _ = fileReadURL(ctx, s.files, r.Thumbnail)
	}
	return out
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
		Thumbnail:    r.Thumbnail,
		UploadedByID: r.UploadedByID,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}
