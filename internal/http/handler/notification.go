package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"fleet/internal/http/dto"
)

// NotificationHandler is a deliberate stub. Nothing in the schema produces
// notifications yet, so the endpoint answers with an empty, well-formed page:
// that lets the bell icon ship and poll now, and turns switching it on later
// into a change of this handler rather than a change of the client.
type NotificationHandler struct{}

func NewNotificationHandler() *NotificationHandler { return &NotificationHandler{} }

// List godoc
//
//	@Summary		List notifications (stub — always empty)
//	@Tags			notifications
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.NotificationListResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Router			/api/v1/notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	c.JSON(http.StatusOK, dto.NotificationListResponse{
		Data:        []dto.NotificationResponse{},
		UnreadCount: 0,
	})
}
