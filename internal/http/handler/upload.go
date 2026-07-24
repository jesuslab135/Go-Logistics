package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/storage"
)

type UploadHandler struct {
	store storage.Storage
}

func NewUploadHandler(store storage.Storage) *UploadHandler {
	return &UploadHandler{store: store}
}

// Upload stores a multipart "file" field and returns its key + public URL. Keys
// are namespaced by tenant and prefixed with a random token so uploads never
// collide or overwrite one another.
// Upload godoc
//
//	@Summary	Upload a file (image) to object storage
//	@Tags		uploads
//	@Security	BearerAuth
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		file	formData	file	true	"File to upload"
//	@Success	201		{object}	storage.Object
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Router		/api/v1/uploads [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	header, err := c.FormFile("file")
	if err != nil {
		apierr.Abort(c, apierr.BadRequest("a multipart 'file' field is required").Wrap(err))
		return
	}

	f, err := header.Open()
	if err != nil {
		apierr.Abort(c, apierr.Internal(err))
		return
	}
	defer f.Close()

	var companyID int64
	if claims, ok := middleware.ClaimsOf(c); ok {
		companyID = claims.CompanyID
	}

	key := objectKey(companyID, header.Filename)
	obj, err := h.store.Save(c.Request.Context(), key, f, header.Header.Get("Content-Type"))
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusCreated, obj)
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
