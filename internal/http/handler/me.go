package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
)

// MeHandler serves /api/v1/me/* endpoints.
type MeHandler struct {
	q *gen.Queries
}

func NewMeHandler(q *gen.Queries) *MeHandler {
	return &MeHandler{q: q}
}

// GetPermissions returns the authenticated employee's merged permissions
// (role permissions + employee flags). The frontend uses this on app load to
// build the sidebar and set route guards.
func (h *MeHandler) GetPermissions(c *gin.Context) {
	claims, ok := middleware.ClaimsOf(c)
	if !ok {
		apierr.Abort(c, apierr.Unauthorized("missing auth claims"))
		return
	}

	employeeID := claims.EmployeeID()
	companyID := claims.CompanyID

	emp, err := h.q.GetEmployee(c.Request.Context(), gen.GetEmployeeParams{
		ID:        employeeID,
		CompanyID: companyID,
	})
	if err != nil {
		apierr.Abort(c, apierr.Map(err))
		return
	}

	resp := dto.MePermissionsResponse{
		UserID:     emp.UserID,
		EmployeeID: emp.ID,
		CompanyID:  companyID,
		Employee: dto.MePermissionsEmployee{
			IsTechnician:      emp.IsTechnician,
			IsVehicleOperator: emp.IsVehicleOperator,
			IsAccountOwner:    emp.IsAccountOwner,
		},
	}

	if emp.RoleID != nil {
		role, err := h.q.GetRole(c.Request.Context(), gen.GetRoleParams{
			ID:        *emp.RoleID,
			CompanyID: companyID,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			apierr.Abort(c, apierr.Map(err))
			return
		}
		if err == nil {
			resp.Role = dto.MePermissionsRole{
				ID:          role.ID,
				Name:        role.Name,
				IsAdmin:     role.IsAdmin,
				Permissions: json.RawMessage(role.Permissions),
			}
		}
	}

	c.JSON(http.StatusOK, resp)
}
