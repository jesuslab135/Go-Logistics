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
//	@Param		purpose	formData	string	false	"asset_photo, company_logo, receipt, signature, inspection_photo, media, photo, document or generic (default). Narrows the accepted media types, and decides whether the object is publicly readable — only asset_photo and company_logo are."
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
	purpose, err := purposeFor(c.PostForm("purpose"))
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	allowed := h.allowedFor(purpose)
	if !allowed[contentType] {
		apierr.Abort(c, apierr.UnsupportedMediaType(
			fmt.Sprintf("file type %q is not allowed", contentType)))
		return
	}

	var companyID int64
	if claims, ok := middleware.ClaimsOf(c); ok {
		companyID = claims.CompanyID
	}

	key := objectKey(purpose.visibility, companyID, header.Filename)
	obj, err := h.store.Save(c.Request.Context(), key, f, contentType)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	// The key is what the caller stores on the owning record; the URL is a
	// convenience for rendering the upload immediately, and for a private object
	// it expires.
	url, expiresAt := fileReadURL(c.Request.Context(), h.store, obj.Key)
	out := dto.UploadResponse{
		Key:         obj.Key,
		URL:         url,
		ExpiresAt:   expiresAt,
		Visibility:  purpose.visibility.String(),
		Size:        obj.Size,
		ContentType: obj.ContentType,
	}
	if thumb, ok := h.saveThumbnail(c, key, contentType, f); ok {
		thumbURL, _ := fileReadURL(c.Request.Context(), h.store, thumb.Key)
		out.ThumbnailKey = thumb.Key
		out.ThumbnailURL = thumbURL
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

// Upload purposes. One endpoint serves photos and documents alike, so the caller
// declares what the file is for. The purpose decides two things: which media
// types are accepted — nothing should store a PDF where a gallery expects an
// image — and whether the object is world-readable.
//
// Visibility is a property of what a file is, not of who uploaded it. A vehicle
// photo and a company logo are branding: they render in plain img tags across
// list screens, and signing them would cost a signature per row to protect
// nothing. Everything else — receipts, signatures, inspection evidence,
// documents, arbitrary media — is a tenant's business records, and is private.
const (
	purposeGeneric = "generic"

	// purposePhoto and purposeDocument predate the public/private split and are
	// kept so existing clients keep working. Both are private: a caller that did
	// not say what the file is for has not established that publishing it is safe.
	purposePhoto    = "photo"
	purposeDocument = "document"

	purposeAssetPhoto      = "asset_photo"
	purposeCompanyLogo     = "company_logo"
	purposeReceipt         = "receipt"
	purposeSignature       = "signature"
	purposeInspectionPhoto = "inspection_photo"
	purposeMedia           = "media"
)

// uploadPurpose is one entry of the purpose vocabulary.
type uploadPurpose struct {
	name       string
	visibility storage.Visibility
	// keep narrows the configured media-type allowlist; nil accepts all of it.
	keep func(mediaType string) bool
}

func imagesOnly(mediaType string) bool { return strings.HasPrefix(mediaType, "image/") }

func nonImagesOnly(mediaType string) bool { return !strings.HasPrefix(mediaType, "image/") }

var uploadPurposes = map[string]uploadPurpose{
	purposeGeneric:         {purposeGeneric, storage.VisibilityPrivate, nil},
	purposePhoto:           {purposePhoto, storage.VisibilityPrivate, imagesOnly},
	purposeDocument:        {purposeDocument, storage.VisibilityPrivate, nonImagesOnly},
	purposeAssetPhoto:      {purposeAssetPhoto, storage.VisibilityPublic, imagesOnly},
	purposeCompanyLogo:     {purposeCompanyLogo, storage.VisibilityPublic, imagesOnly},
	purposeReceipt:         {purposeReceipt, storage.VisibilityPrivate, nil},
	purposeSignature:       {purposeSignature, storage.VisibilityPrivate, imagesOnly},
	purposeInspectionPhoto: {purposeInspectionPhoto, storage.VisibilityPrivate, imagesOnly},
	purposeMedia:           {purposeMedia, storage.VisibilityPrivate, nil},
}

// purposeNames lists the vocabulary for error messages, in a stable order.
var purposeNames = []string{
	purposeGeneric, purposePhoto, purposeDocument,
	purposeAssetPhoto, purposeCompanyLogo, purposeReceipt,
	purposeSignature, purposeInspectionPhoto, purposeMedia,
}

// purposeFor resolves the declared purpose. An unrecognised one is refused
// rather than silently treated as generic, since that would quietly widen both
// the media-type policy and the visibility the caller asked for.
func purposeFor(declared string) (uploadPurpose, error) {
	name := strings.ToLower(strings.TrimSpace(declared))
	if name == "" {
		name = purposeGeneric
	}
	p, ok := uploadPurposes[name]
	if !ok {
		return uploadPurpose{}, apierr.New(http.StatusBadRequest, "invalid_upload_purpose",
			fmt.Sprintf("purpose must be one of %s", strings.Join(purposeNames, ", ")))
	}
	return p, nil
}

// allowedFor narrows the configured allowlist to the declared purpose.
func (h *UploadHandler) allowedFor(p uploadPurpose) map[string]bool {
	if p.keep == nil {
		return h.allowed
	}
	return h.subset(p.keep)
}

// subset keeps a purpose from widening the configured allowlist: it can only
// ever remove types from it, never add one.
func (h *UploadHandler) subset(keep func(string) bool) map[string]bool {
	out := make(map[string]bool, len(h.allowed))
	for mediaType, ok := range h.allowed {
		if ok && keep(mediaType) {
			out[mediaType] = true
		}
	}
	return out
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

// objectKey namespaces an upload by visibility and tenant, and prefixes a random
// token so uploads never collide or overwrite one another. The visibility
// segment is what lets every later operation route the object to the right
// bucket from the key alone.
func objectKey(vis storage.Visibility, companyID int64, filename string) string {
	token := make([]byte, 8)
	_, _ = rand.Read(token)
	return fmt.Sprintf("%s%d/%s-%s", vis.KeyPrefix(), companyID, hex.EncodeToString(token), safeName(filename))
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
