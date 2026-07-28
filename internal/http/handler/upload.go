package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/imaging"
	"fleet/internal/platform/storage"
)

const (
	defaultUploadMaxBytes     = 5 << 20
	defaultUploadAllowedTypes = "image/jpeg,image/png,image/gif,image/webp,application/pdf"

	// http.DetectContentType never looks past this many bytes.
	sniffLen = 512
)

type UploadHandler struct {
	store    storage.Storage
	maxBytes int64
	allowed  map[string]bool
	thumbDim int
	log      *slog.Logger
}

func NewUploadHandler(store storage.Storage, log *slog.Logger) *UploadHandler {
	if log == nil {
		log = slog.Default()
	}
	return &UploadHandler{
		store:    store,
		maxBytes: uploadMaxBytes(),
		allowed:  uploadAllowedTypes(),
		thumbDim: thumbnailMaxDim(),
		log:      log,
	}
}

// Upload stores a multipart "file" field and returns its key + public URL. Keys
// are namespaced by tenant and prefixed with a random token so uploads never
// collide or overwrite one another.
//
// The declared Content-Type is deliberately ignored. It is client-supplied, and
// the bucket is public-read, so trusting it would let a caller have our own
// domain serve arbitrary content under a type of their choosing. The type is
// sniffed from the bytes instead, and only then checked against the allowlist.
// Upload godoc
//
//	@Summary	Upload a file (image) to object storage
//	@Tags		uploads
//	@Security	BearerAuth
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		file	formData	file	true	"File to upload"
//	@Success	201		{object}	dto.UploadResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Failure	413		{object}	dto.ErrorResponse
//	@Failure	415		{object}	dto.ErrorResponse
//	@Router		/api/v1/uploads [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	// Bound the body before Gin buffers it, so an oversized upload is refused
	// mid-stream instead of after being spooled to a temp file.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxBytes)

	header, err := c.FormFile("file")
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			apierr.Abort(c, h.tooLargeErr().Wrap(err))
			return
		}
		apierr.Abort(c, apierr.BadRequest("a multipart 'file' field is required").Wrap(err))
		return
	}
	if header.Size > h.maxBytes {
		apierr.Abort(c, h.tooLargeErr())
		return
	}

	f, err := header.Open()
	if err != nil {
		apierr.Abort(c, apierr.Internal(err))
		return
	}
	defer f.Close()

	contentType, err := sniffContentType(f)
	if err != nil {
		apierr.Abort(c, apierr.BadRequest("could not read the uploaded file").Wrap(err))
		return
	}
	if !h.allowed[contentType] {
		apierr.Abort(c, apierr.UnsupportedMediaType(
			fmt.Sprintf("file type %q is not allowed", contentType)))
		return
	}

	var companyID int64
	if claims, ok := middleware.ClaimsOf(c); ok {
		companyID = claims.CompanyID
	}

	key := objectKey(companyID, header.Filename)
	obj, err := h.store.Save(c.Request.Context(), key, f, contentType)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	out := dto.UploadResponse{
		Key:         obj.Key,
		URL:         obj.URL,
		Size:        obj.Size,
		ContentType: obj.ContentType,
	}
	if thumb, ok := h.saveThumbnail(c, key, contentType, f); ok {
		out.ThumbnailKey = thumb.Key
		out.ThumbnailURL = thumb.URL
	}

	c.JSON(http.StatusCreated, out)
}

// saveThumbnail derives and stores a thumbnail beside the original. A failure
// here is logged and swallowed: the upload itself succeeded, and refusing it
// because a preview could not be produced would lose the user's file over a
// cosmetic feature. Callers see the missing thumbnail fields and fall back to
// the original.
func (h *UploadHandler) saveThumbnail(c *gin.Context, key, contentType string, f io.ReadSeeker) (storage.Object, bool) {
	if !imaging.Thumbnailable(contentType) {
		return storage.Object{}, false
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		h.log.WarnContext(c.Request.Context(), "thumbnail skipped: rewind failed", "key", key, "error", err)
		return storage.Object{}, false
	}

	data, err := imaging.Thumbnail(f, h.thumbDim)
	if err != nil {
		h.log.WarnContext(c.Request.Context(), "thumbnail skipped: generation failed", "key", key, "error", err)
		return storage.Object{}, false
	}

	obj, err := h.store.Save(c.Request.Context(), imaging.ThumbKey(key), bytes.NewReader(data), imaging.ThumbContentType)
	if err != nil {
		h.log.WarnContext(c.Request.Context(), "thumbnail skipped: store failed", "key", key, "error", err)
		return storage.Object{}, false
	}
	return obj, true
}

func (h *UploadHandler) tooLargeErr() *apierr.Error {
	return apierr.PayloadTooLarge(fmt.Sprintf("file exceeds the %d byte limit", h.maxBytes))
}

// sniffContentType detects the real media type from the leading bytes and
// rewinds, so the caller can still stream the whole file to storage.
func sniffContentType(f io.ReadSeeker) (string, error) {
	buf := make([]byte, sniffLen)
	n, err := io.ReadFull(f, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	// DetectContentType appends a charset for text types; match on the media
	// type alone.
	mediaType, _, _ := strings.Cut(http.DetectContentType(buf[:n]), ";")
	return strings.ToLower(strings.TrimSpace(mediaType)), nil
}

func thumbnailMaxDim() int {
	if v, err := strconv.Atoi(os.Getenv("UPLOAD_THUMBNAIL_MAX_DIM")); err == nil && v > 0 {
		return v
	}
	return imaging.DefaultMaxDim
}

func uploadMaxBytes() int64 {
	if v, err := strconv.ParseInt(os.Getenv("UPLOAD_MAX_BYTES"), 10, 64); err == nil && v > 0 {
		return v
	}
	return defaultUploadMaxBytes
}

// uploadAllowedTypes is an allowlist of sniffable media types. Django also
// accepted office documents and CSV, but those cannot be told apart by content
// sniffing (docx/xlsx are zip archives), so they are opt-in via the env var
// rather than allowed by default.
func uploadAllowedTypes() map[string]bool {
	raw := os.Getenv("UPLOAD_ALLOWED_TYPES")
	if strings.TrimSpace(raw) == "" {
		raw = defaultUploadAllowedTypes
	}
	allowed := make(map[string]bool)
	for _, t := range strings.Split(raw, ",") {
		if t = strings.ToLower(strings.TrimSpace(t)); t != "" {
			allowed[t] = true
		}
	}
	return allowed
}

func objectKey(companyID int64, filename string) string {
	token := make([]byte, 8)
	_, _ = rand.Read(token)
	return fmt.Sprintf("uploads/%d/%s-%s", companyID, hex.EncodeToString(token), safeName(filename))
}

func safeName(filename string) string {
	base := filepath.Base(strings.ReplaceAll(filename, "\\", "/"))
	base = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, base)
	if base == "" || base == "." {
		return "file"
	}
	return base
}
