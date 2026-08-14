package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/paginate"
)

// Notification kinds. Each is produced by a specific event; the frontend keys
// its icon and copy off these.
const (
	NotificationTireAssignmentPending  = "tire_assignment_pending"
	NotificationTireAssignmentApproved = "tire_assignment_approved"
)

type NotificationHandler struct{ q *gen.Queries }

func NewNotificationHandler(q *gen.Queries) *NotificationHandler {
	return &NotificationHandler{q: q}
}

// List godoc
//
//	@Summary		List the caller's notifications
//	@Description	Scoped to the authenticated employee within the active company, newest first. The unread count describes the whole inbox, not just the returned page.
//	@Tags			notifications
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int	false	"Page size"
//	@Param			offset	query		int	false	"Offset"
//	@Success		200		{object}	dto.NotificationListResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Router			/api/v1/notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)
	employee := middleware.EmployeeFromContext(ctx)
	company := middleware.CompanyFromContext(ctx)

	rows, err := h.q.ListNotifications(ctx, gen.ListNotificationsParams{
		EmployeeID: employee,
		CompanyID:  company,
		Limit:      int32(p.Limit),
		Offset:     int32(p.Offset),
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	unread, err := h.q.CountUnreadNotifications(ctx, gen.CountUnreadNotificationsParams{
		EmployeeID: employee,
		CompanyID:  company,
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	out := make([]dto.NotificationResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, toNotificationResponse(r))
	}
	c.JSON(http.StatusOK, dto.NotificationListResponse{Data: out, UnreadCount: unread})
}

// MarkRead godoc
//
//	@Summary		Mark a notification read
//	@Description	Idempotent: marking an already-read notification keeps its original read_at. A caller can only mark their own, so another employee's id is indistinguishable from a missing one.
//	@Tags			notifications
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Notification id"
//	@Success		200	{object}	dto.NotificationResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/notifications/{id}/read [post]
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		apierr.Abort(c, apierr.BadRequest("invalid id"))
		return
	}

	ctx := c.Request.Context()
	now := time.Now().UTC()
	r, err := h.q.MarkNotificationRead(ctx, gen.MarkNotificationReadParams{
		ID:         id,
		EmployeeID: middleware.EmployeeFromContext(ctx),
		CompanyID:  middleware.CompanyFromContext(ctx),
		ReadAt:     &now,
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, toNotificationResponse(r))
}

func toNotificationResponse(r gen.Notification) dto.NotificationResponse {
	return dto.NotificationResponse{
		ID:        r.ID,
		Kind:      r.Kind,
		Title:     r.Title,
		Body:      r.Body,
		URL:       r.Url,
		ReadAt:    r.ReadAt,
		CreatedAt: r.CreatedAt,
	}
}
