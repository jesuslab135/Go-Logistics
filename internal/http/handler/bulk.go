package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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

// importMaxRowsDefault is the default for IMPORT_MAX_ROWS. sheet.maxDataRows
// (5001: this many data rows plus the header) bounds how far a template's
// drop-down lists reach, and the two must move together - raise this without
// raising that and every row past 5001 silently loses its drop-downs.
const importMaxRowsDefault = 5000

// exportMaxRowsDefault is the default for EXPORT_MAX_ROWS. It deliberately
// equals importMaxRowsDefault: export is documented as the way to seed another
// company, so a cap larger than the import's would hand the user a file the
// import then refuses - discovered only after the download. Raise one and
// raise the other, or exports stop round-tripping.
const exportMaxRowsDefault = importMaxRowsDefault

// maxConcurrentImports is how many imports may run at once, and
// maxConcurrentExports how many exports.
//
// bulk.Run holds one pooled connection for an entire import - up to
// IMPORT_MAX_ROWS rows in a single transaction - where every other handler in
// this API holds one for milliseconds. With a pool ceiling of 20
// (pool_max_conns in the DSN; see the README's environment section) a few
// simultaneous imports must not be able to starve every other request, and the
// server sets only ReadHeaderTimeout, so anything that blocks in Acquire
// blocks indefinitely. Past the limit an import is refused with 429 rather
// than queued behind a wait nothing will break.
//
// Exports are bounded for memory rather than connections: an export buffers
// the whole result set before writing it, and neither compose file sets a
// memory limit on the API container.
const (
	maxConcurrentImports = 2
	maxConcurrentExports = 2
)

// acquire takes one slot without waiting, reporting whether it got one. The
// caller releases it with the returned func. Refusing beats queueing here: a
// caller who waits behind a multi-minute import has already timed out.
func acquire(slots chan struct{}) (func(), bool) {
	select {
	case slots <- struct{}{}:
		return func() { <-slots }, true
	default:
		return nil, false
	}
}

// BulkHandler serves spreadsheet imports and their templates. Routes are
// registered on each section's own group, so the module permission that gates
// creating a record also gates importing it.
type BulkHandler struct {
	pool     *pgxpool.Pool
	maxBytes int64
	limits   bulk.Limits
	slots    chan struct{}
}

func NewBulkHandler(pool *pgxpool.Pool) *BulkHandler {
	return &BulkHandler{
		pool:     pool,
		maxBytes: bulkEnvInt("IMPORT_MAX_BYTES", 10<<20),
		limits:   bulk.Limits{MaxRows: int(bulkEnvInt("IMPORT_MAX_ROWS", importMaxRowsDefault))},
		slots:    make(chan struct{}, maxConcurrentImports),
	}
}

// bulkEnvInt reads a positive integer from the environment, falling back to
// the default when the variable is absent or empty. A variable that is present
// but unusable - a typo, a negative, or the 0 an operator may well write
// meaning "no limit" - also falls back, but says so first: silently applying
// 5000 rows to someone who asked for something else leaves the limit to be
// discovered by a confusing rejection much later.
func bulkEnvInt(key string, fallback int64) int64 {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}
	v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || v <= 0 {
		slog.Warn("bulk: ignoring unusable limit, using the default",
			"variable", key, "value", raw, "default", fallback)
		return fallback
	}
	return v
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

		// Taken around the import alone. Reading and parsing the upload holds
		// no database connection, so a slot spent covering that would turn
		// callers away for nothing.
		release, ok := acquire(h.slots)
		if !ok {
			apierr.Abort(c, apierr.TooManyRequests(fmt.Sprintf(
				"%d imports are already running; wait for one to finish and try again", maxConcurrentImports)))
			return
		}
		defer release()

		report, err := bulk.Run(c.Request.Context(), h.pool, imp, rows, dryRun, h.limits)
		if err != nil {
			apierr.Abort(c, err)
			return
		}
		if len(report.Errors) > 0 {
			apierr.Abort(c, apierr.New(http.StatusUnprocessableEntity, "import_failed",
				fmt.Sprintf("%d problem(s) found; nothing was imported", report.ErrorCount)).WithDetails(report))
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

// exportPageSize is how many rows the exporter asks the list endpoint for at
// a time.
const exportPageSize = 100

// Exporter writes a list to a file by calling the section's own list endpoint
// page by page, so every filter and permission of that list applies unchanged.
type Exporter struct {
	engine  http.Handler
	maxRows int
	slots   chan struct{}
}

func NewExporter(engine http.Handler) *Exporter {
	return &Exporter{
		engine:  engine,
		maxRows: int(bulkEnvInt("EXPORT_MAX_ROWS", exportMaxRowsDefault)),
		slots:   make(chan struct{}, maxConcurrentExports),
	}
}

func (e *Exporter) Handler(listPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		format := c.DefaultQuery("format", "xlsx")
		if format != "xlsx" && format != "csv" {
			apierr.Abort(c, apierr.BadRequest(`format must be "xlsx" or "csv"`))
			return
		}
		// Held for the whole export: every page fetched accumulates in the
		// table below, so it is concurrency, not any single request, that
		// decides peak memory.
		release, ok := acquire(e.slots)
		if !ok {
			apierr.Abort(c, apierr.TooManyRequests(fmt.Sprintf(
				"%d exports are already running; wait for one to finish and try again", maxConcurrentExports)))
			return
		}
		defer release()

		query := c.Request.URL.Query()
		for _, k := range []string{"format", "page", "page_size"} {
			query.Del(k)
		}

		var table bulk.Table
		fetched, truncated := 0, false
		for fetched < e.maxRows {
			// Clamped to what is left under the cap, so the cap is exact
			// rather than being able to overshoot by up to a whole page.
			limit := e.maxRows - fetched
			if limit > exportPageSize {
				limit = exportPageSize
			}
			query.Set("limit", strconv.Itoa(limit))
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
			if len(page.Data) == 0 || !page.HasNext {
				break
			}
			if fetched >= e.maxRows {
				truncated = true
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
			// Streamed, not WriteXLSX: an export is a single sheet with no
			// data validation, so it does not need what WriteXLSX offers, and
			// building 800k cells as an object graph first is what turns a
			// large export into an OOM kill of the whole API.
			err = sheet.WriteXLSXStream(c.Writer, sheet.DataSheet, table.Records())
		}
		if err != nil {
			_ = c.Error(err)
		}
	}
}
