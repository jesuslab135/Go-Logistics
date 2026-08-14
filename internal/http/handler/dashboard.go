package handler

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
)

const defaultDashboardUpcomingDays = 30

type DashboardHandler struct {
	q *gen.Queries
	// upcomingDays is how far ahead a service reminder counts as "upcoming".
	upcomingDays int32
}

func NewDashboardHandler(q *gen.Queries) *DashboardHandler {
	return &DashboardHandler{q: q, upcomingDays: dashboardUpcomingDays()}
}

// Stats godoc
//
//	@Summary		Aggregated dashboard KPIs
//	@Description	Current-state counters for the caller's company. pending_inspections counts submissions with failed items and no issue raised from them; low_stock_parts counts inventory rows (a part at one location) at or below their reorder point, not distinct parts. upcoming_days reports the window upcoming_reminders was counted over.
//	@Tags			dashboard
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.DashboardStatsResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Router			/api/v1/dashboard/stats [get]
func (h *DashboardHandler) Stats(c *gin.Context) {
	row, err := h.q.GetDashboardStats(c.Request.Context(), gen.GetDashboardStatsParams{
		CompanyID:    middleware.CompanyFromContext(c.Request.Context()),
		UpcomingDays: h.upcomingDays,
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.DashboardStatsResponse{
		TotalAssets:        row.TotalAssets,
		ActiveWorkOrders:   row.ActiveWorkOrders,
		OverdueIssues:      row.OverdueIssues,
		UpcomingReminders:  row.UpcomingReminders,
		LowStockParts:      row.LowStockParts,
		PendingInspections: row.PendingInspections,
		UpcomingDays:       h.upcomingDays,
	})
}

func dashboardUpcomingDays() int32 {
	if v, err := strconv.ParseInt(os.Getenv("DASHBOARD_UPCOMING_DAYS"), 10, 32); err == nil && v > 0 {
		return int32(v)
	}
	return defaultDashboardUpcomingDays
}
