package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/platform/apierr"
	"fleet/internal/platform/bulk"
	"fleet/internal/platform/sheet"
)

const xlsxContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

// BulkHandler serves spreadsheet imports and their templates. Routes are
// registered on each section's own group, so the module permission that gates
// creating a record also gates importing it.
type BulkHandler struct {
	pool     *pgxpool.Pool
	maxBytes int64
	limits   bulk.Limits
}

func NewBulkHandler(pool *pgxpool.Pool) *BulkHandler {
	return &BulkHandler{
		pool:     pool,
		maxBytes: bulkEnvInt("IMPORT_MAX_BYTES", 10<<20),
		limits:   bulk.Limits{MaxRows: int(bulkEnvInt("IMPORT_MAX_ROWS", 5000))},
	}
}

func bulkEnvInt(key string, fallback int64) int64 {
	if v, err := strconv.ParseInt(os.Getenv(key), 10, 64); err == nil && v > 0 {
		return v
	}
	return fallback
}

// Register wires POST path/import and GET path/import/template.
func (h *BulkHandler) Register(g gin.IRouter, path string, imp bulk.Importer) {
	g.POST(path+"/import", h.importRows(imp))
	g.GET(path+"/import/template", h.template(imp))
}

func (h *BulkHandler) importRows(imp bulk.Importer) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxBytes)
		fh, err := c.FormFile("file")
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				apierr.Abort(c, apierr.PayloadTooLarge(fmt.Sprintf("the file is larger than %d bytes", h.maxBytes)))
				return
			}
			apierr.Abort(c, apierr.BadRequest(`send the spreadsheet as the multipart form field "file"`))
			return
		}
		f, err := fh.Open()
		if err != nil {
			apierr.Abort(c, apierr.BadRequest("the file could not be opened"))
			return
		}
		defer f.Close()

		rows, err := sheet.Read(f, fh.Filename)
		if errors.Is(err, sheet.ErrUnsupported) {
			apierr.Abort(c, apierr.UnsupportedMediaType(err.Error()))
			return
		}
		if err != nil {
			apierr.Abort(c, apierr.BadRequest("the file could not be read: "+err.Error()))
			return
		}

		dryRun := c.Query("dry_run") == "true" || c.Query("dry_run") == "1"
		report, err := bulk.Run(c.Request.Context(), h.pool, imp, rows, dryRun, h.limits)
		if err != nil {
			apierr.Abort(c, err)
			return
		}
		if len(report.Errors) > 0 {
			apierr.Abort(c, apierr.New(http.StatusUnprocessableEntity, "import_failed",
				fmt.Sprintf("%d problem(s) found; nothing was imported", len(report.Errors))).WithDetails(report))
			return
		}
		status := http.StatusCreated
		if dryRun {
			status = http.StatusOK
		}
		c.JSON(status, report)
	}
}

func (h *BulkHandler) template(imp bulk.Importer) gin.HandlerFunc {
	return func(c *gin.Context) {
		sheets, err := bulk.Template(c.Request.Context(), imp)
		if err != nil {
			apierr.Abort(c, err)
			return
		}
		c.Header("Content-Type", xlsxContentType)
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-plantilla.xlsx"`, imp.Resource()))
		if err := sheet.WriteXLSX(c.Writer, sheets); err != nil {
			_ = c.Error(err)
		}
	}
}
