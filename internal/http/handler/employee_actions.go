package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"fleet/internal/auth"
	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
)

type EmployeeActionHandler struct {
	q *gen.Queries
}

func NewEmployeeActionHandler(q *gen.Queries) *EmployeeActionHandler {
	return &EmployeeActionHandler{q: q}
}

// SetPassword godoc
//
//	@Summary		Set an employee's password
//	@Description	Lets a company administrator provision login credentials without running the CLI. password_hash is never accepted or returned.
//	@Tags			employees
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path	int								true	"Employee id"
//	@Param			password	body	dto.SetEmployeePasswordRequest	true	"New password"
//	@Success		204			"password updated"
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		403			{object}	dto.ErrorResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Failure		422			{object}	dto.ErrorResponse
//	@Router			/api/v1/employees/{id}/set-password [post]
func (h *EmployeeActionHandler) SetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		apierr.Abort(c, apierr.BadRequest("invalid id"))
		return
	}

	var req dto.SetEmployeePasswordRequest
	if err := bindJSONValidated(c, &req); err != nil {
		apierr.Abort(c, err)
		return
	}

	ctx := c.Request.Context()
	// The lookup is scoped by employee_companies membership, so an employee of
	// another company is indistinguishable from a missing one: 404, not 403.
	if _, err := h.q.GetEmployee(ctx, gen.GetEmployeeParams{
		ID:        id,
		CompanyID: middleware.CompanyFromContext(ctx),
	}); err != nil {
		apierr.Abort(c, err)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if err := h.q.UpdateEmployeePassword(ctx, gen.UpdateEmployeePasswordParams{
		ID:           id,
		PasswordHash: hash,
		UpdatedAt:    time.Now().UTC(),
	}); err != nil {
		apierr.Abort(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
