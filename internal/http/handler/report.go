package handler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/mail"
	"fleet/internal/platform/paginate"
	"fleet/internal/platform/pdf"
	"fleet/internal/platform/reports"
	"fleet/internal/platform/storage"
)

// maxLogoBytes bounds what is read into memory for a PDF header. Uploads are
// already capped by UPLOAD_MAX_BYTES; this is the second fence, for objects
// that reached the bucket some other way.
const maxLogoBytes = 5 << 20

// NewLogoLoader returns the function that reads a company's logo for a PDF.
// It lives here, not in the scheduler, because resolving a stored file
// reference safely (fileKeyForCompany) is this package's knowledge.
//
// Every failure yields nil: a report without a logo is still the report.
func NewLogoLoader(files storage.Storage, logger *slog.Logger) func(ctx context.Context, c reports.Company) []byte {
	return func(ctx context.Context, c reports.Company) []byte {
		if files == nil || c.Logo == nil {
			return nil
		}
		key, ok := fileKeyForCompany(files, *c.Logo, c.ID)
		if !ok {
			return nil
		}
		rc, err := files.Open(ctx, key)
		if err != nil {
			logger.Warn("report logo: open", "company_id", c.ID, "error", err)
			return nil
		}
		defer rc.Close()
		data, err := io.ReadAll(io.LimitReader(rc, maxLogoBytes+1))
		if err != nil || len(data) > maxLogoBytes {
			logger.Warn("report logo: unreadable or too large", "company_id", c.ID, "error", err)
			return nil
		}
		return data
	}
}

type ReportHandler struct {
	q    *gen.Queries
	logo func(ctx context.Context, c reports.Company) []byte
	now  func() time.Time
}

func NewReportHandler(q *gen.Queries, files storage.Storage, logger *slog.Logger) *ReportHandler {
	return &ReportHandler{q: q, logo: NewLogoLoader(files, logger), now: time.Now}
}

// ListRecipients godoc
//
//	@Summary		List who receives the scheduled reports
//	@Description	The addresses the active company's scheduled reports are emailed to, optionally for one report.
//	@Tags			reports
//	@Produce		json
//	@Security		BearerAuth
//	@Param			report_kind	query		string	false	"fuel_weekly or maintenance_monthly"
//	@Success		200			{object}	dto.ReportRecipientListResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		403			{object}	dto.ErrorResponse
//	@Router			/api/v1/reports/recipients [get]
func (h *ReportHandler) ListRecipients(c *gin.Context) {
	ctx := c.Request.Context()

	var kind *string
	if v := c.Query("report_kind"); v != "" {
		if _, ok := reports.ParseKind(v); !ok {
			apierr.Abort(c, apierr.BadRequest("report_kind must be fuel_weekly or maintenance_monthly"))
			return
		}
		kind = &v
	}

	rows, err := h.q.ListReportRecipients(ctx, gen.ListReportRecipientsParams{
		CompanyID:  middleware.CompanyFromContext(ctx),
		ReportKind: kind,
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	out := make([]dto.ReportRecipientResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, toReportRecipientResponse(r))
	}
	c.JSON(http.StatusOK, dto.ReportRecipientListResponse{Data: out})
}

// CreateRecipient godoc
//
//	@Summary		Add a report recipient
//	@Description	The address must be a bare email address. The same address cannot be listed twice for one report; letter case is ignored.
//	@Tags			reports
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.CreateReportRecipientRequest	true	"body"
//	@Success		201		{object}	dto.ReportRecipientResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/api/v1/reports/recipients [post]
func (h *ReportHandler) CreateRecipient(c *gin.Context) {
	var in dto.CreateReportRecipientRequest
	if err := bindJSONValidated(c, &in); err != nil {
		apierr.Abort(c, err)
		return
	}

	kind, ok := reports.ParseKind(in.ReportKind)
	if !ok {
		apierr.Abort(c, apierr.Validation(map[string]string{"report_kind": "must be fuel_weekly or maintenance_monthly"}))
		return
	}
	// The address ends up in a message header. CleanAddress refuses a display
	// name and a line break, which would otherwise add headers to that message.
	email, err := mail.CleanAddress(in.Email)
	if err != nil {
		apierr.Abort(c, apierr.Validation(map[string]string{"email": "must be a single email address"}))
		return
	}

	ctx := c.Request.Context()
	row, err := h.q.CreateReportRecipient(ctx, gen.CreateReportRecipientParams{
		CompanyID:  middleware.CompanyFromContext(ctx),
		ReportKind: string(kind),
		Email:      email,
	})
	if err != nil {
		// A duplicate trips uq_report_recipient_company_kind_email, which
		// apierr.Map turns into 409.
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusCreated, toReportRecipientResponse(row))
}

// DeleteRecipient godoc
//
//	@Summary		Remove a report recipient
//	@Tags			reports
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Recipient id"
//	@Success		204
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/reports/recipients/{id} [delete]
func (h *ReportHandler) DeleteRecipient(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		apierr.Abort(c, apierr.BadRequest("invalid id"))
		return
	}
	ctx := c.Request.Context()
	n, err := h.q.DeleteReportRecipient(ctx, gen.DeleteReportRecipientParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if n == 0 {
		apierr.Abort(c, apierr.NotFound("recipient not found"))
		return
	}
	c.Status(http.StatusNoContent)
}

// ListRuns godoc
//
//	@Summary		List the scheduled report runs
//	@Description	What the scheduler did for the active company, newest first: sent, skipped, or failed and why.
//	@Tags			reports
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit		query		int	false	"Page size"
//	@Param			offset		query		int	false	"Offset"
//	@Success		200			{object}	paginate.Page[dto.ReportRunResponse]
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		403			{object}	dto.ErrorResponse
//	@Router			/api/v1/reports/runs [get]
func (h *ReportHandler) ListRuns(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)
	company := middleware.CompanyFromContext(ctx)

	rows, err := h.q.ListReportRuns(ctx, gen.ListReportRunsParams{
		CompanyID: company,
		Lim:       int32(p.Limit),
		Off:       int32(p.Offset),
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	total, err := h.q.CountReportRuns(ctx, company)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	out := make([]dto.ReportRunResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.ReportRunResponse{
			ID:          r.ID,
			ReportKind:  r.ReportKind,
			PeriodStart: r.PeriodStart.Format("2006-01-02"),
			Status:      r.Status,
			Attempts:    r.Attempts,
			Error:       r.Error,
			Recipients:  r.Recipients,
			StartedAt:   r.StartedAt,
			FinishedAt:  r.FinishedAt,
		})
	}
	c.JSON(http.StatusOK, paginate.NewPage(out, total, p))
}

// FuelWeekly godoc
//
//	@Summary		Download the weekly fuel report
//	@Description	The same PDF the scheduler emails. week is any date inside the week wanted, read in the company's timezone; without it, the last full week. A week with no fuel entries still returns a PDF, which says so.
//	@Tags			reports
//	@Produce		application/pdf
//	@Security		BearerAuth
//	@Param			week	query		string	false	"Any date in the week, YYYY-MM-DD"
//	@Success		200		{file}		binary
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Router			/api/v1/reports/fuel/weekly [get]
func (h *ReportHandler) FuelWeekly(c *gin.Context) {
	h.download(c, reports.FuelWeekly, "week", "2006-01-02", "week must be a date, YYYY-MM-DD")
}

// MaintenanceMonthly godoc
//
//	@Summary		Download the monthly maintenance cost report
//	@Description	The same PDF the scheduler emails. Without month, the last full month. A month with no completed jobs still returns a PDF, which says so.
//	@Tags			reports
//	@Produce		application/pdf
//	@Security		BearerAuth
//	@Param			month	query		string	false	"The month, YYYY-MM"
//	@Success		200		{file}		binary
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Router			/api/v1/reports/maintenance/monthly [get]
func (h *ReportHandler) MaintenanceMonthly(c *gin.Context) {
	h.download(c, reports.MaintenanceMonthly, "month", "2006-01", "month must be YYYY-MM")
}

func (h *ReportHandler) download(c *gin.Context, kind reports.Kind, param, layout, invalid string) {
	ctx := c.Request.Context()

	row, err := h.q.GetReportCompany(ctx, middleware.CompanyFromContext(ctx))
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	company := reports.Company{ID: row.ID, Name: row.Name, Logo: row.Logo, Timezone: row.Timezone, Currency: row.Currency}
	// The fallback location is good enough for a download; the scheduler is
	// where an unknown timezone gets logged.
	loc, _ := reports.LoadLocation(company.Timezone)

	now := h.now()
	// Without the parameter: the last period that has ended, whatever the hour.
	period := reports.PeriodOf(kind, now, loc).Previous(kind)
	if v := strings.TrimSpace(c.Query(param)); v != "" {
		at, err := time.ParseInLocation(layout, v, loc)
		if err != nil {
			apierr.Abort(c, apierr.BadRequest(invalid))
			return
		}
		period = reports.PeriodOf(kind, at, loc)
	}

	doc, err := pdf.Render(ctx, h.q, kind, company, period, h.logo(ctx, company), now)
	if err != nil {
		apierr.Abort(c, apierr.Internal(err))
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, doc.Filename))
	c.Data(http.StatusOK, "application/pdf", doc.Data)
}

func toReportRecipientResponse(r gen.ReportRecipient) dto.ReportRecipientResponse {
	return dto.ReportRecipientResponse{
		ID:         r.ID,
		ReportKind: r.ReportKind,
		Email:      r.Email,
		IsActive:   r.IsActive,
		CreatedAt:  r.CreatedAt,
	}
}
