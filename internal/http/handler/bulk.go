package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"

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

// parseDryRun reads ?dry_run= strictly: absent means false, and anything
// present that strconv.ParseBool does not recognize (a typo like "yes", or
// unrelated junk) is an error rather than being silently treated as false.
// A caller testing an import before committing to it is exactly the person
// who must never have a typo turn their dry run into a real write.
func parseDryRun(c *gin.Context) (bool, error) {
	raw := strings.TrimSpace(strings.ToLower(c.Query("dry_run")))
	if raw == "" {
		return false, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("dry_run must be a boolean (true/false/1/0), got %q", raw)
	}
	return v, nil
}

// Register wires POST path/import and GET path/import/template.
func (h *BulkHandler) Register(g gin.IRouter, path string, imp bulk.Importer) {
	g.POST(path+"/import", h.importRows(imp))
	g.GET(path+"/import/template", h.template(imp))
}

func (h *BulkHandler) importRows(imp bulk.Importer) gin.HandlerFunc {
	return func(c *gin.Context) {
		dryRun, err := parseDryRun(c)
		if err != nil {
			apierr.Abort(c, apierr.BadRequest(err.Error()))
			return
		}

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

// Exporter writes a list to a file by calling the section's own list endpoint
// page by page, so every filter and permission of that list applies unchanged.
type Exporter struct {
	engine  http.Handler
	maxRows int
}

func NewExporter(engine http.Handler) *Exporter {
	return &Exporter{engine: engine, maxRows: int(bulkEnvInt("EXPORT_MAX_ROWS", 20000))}
}

func (e *Exporter) Handler(listPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		format := c.DefaultQuery("format", "xlsx")
		if format != "xlsx" && format != "csv" {
			apierr.Abort(c, apierr.BadRequest(`format must be "xlsx" or "csv"`))
			return
		}
		query := c.Request.URL.Query()
		for _, k := range []string{"format", "page", "page_size"} {
			query.Del(k)
		}

		var table bulk.Table
		fetched, truncated := 0, false
		for {
			if fetched >= e.maxRows {
				truncated = true
				break
			}
			query.Set("limit", "100")
			query.Set("offset", strconv.Itoa(fetched))
			req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, "/api/v1"+listPath+"?"+query.Encode(), nil)
			if err != nil {
				apierr.Abort(c, apierr.Internal(err))
				return
			}
			req.Header.Set("Authorization", c.GetHeader("Authorization"))
			rec := httptest.NewRecorder()
			e.engine.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				c.Data(rec.Code, "application/json", rec.Body.Bytes())
				c.Abort()
				return
			}
			var page struct {
				Data    []json.RawMessage `json:"data"`
				HasNext bool              `json:"has_next"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
				apierr.Abort(c, apierr.Internal(err))
				return
			}
			for _, raw := range page.Data {
				if err := table.Add(raw); err != nil {
					apierr.Abort(c, apierr.Internal(err))
					return
				}
			}
			fetched += len(page.Data)
			if !page.HasNext || len(page.Data) == 0 {
				break
			}
		}

		if truncated {
			c.Header("X-Export-Truncated", "true")
		}
		filename := strings.Trim(strings.ReplaceAll(listPath, "/", "-"), "-") + "." + format
		c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
		var err error
		if format == "csv" {
			c.Header("Content-Type", "text/csv; charset=utf-8")
			err = sheet.WriteCSV(c.Writer, table.Records())
		} else {
			c.Header("Content-Type", xlsxContentType)
			err = sheet.WriteXLSX(c.Writer, []sheet.Sheet{{Name: sheet.DataSheet, Rows: table.Records(), FreezeHeader: true}})
		}
		if err != nil {
			_ = c.Error(err)
		}
	}
}
