package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/paginate"
)

// AdminCompanyHandler serves the cross-company /admin/companies routes: owner
// read/set here, and the roles and work-order-statuses lists. Every route is
// addressed by explicit company id rather than the caller's own membership,
// which is what the /api/v1/companies namespace already covers.
type AdminCompanyHandler struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewAdminCompanyHandler(q *gen.Queries, pool *pgxpool.Pool) *AdminCompanyHandler {
	return &AdminCompanyHandler{q: q, pool: pool}
}

// adminCompanyParam parses the :id path param shared by every
// AdminCompanyHandler method.
func adminCompanyParam(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, apierr.BadRequest("invalid id")
	}
	return id, nil
}

func toCompanyOwnerResponse(row gen.GetCompanyOwnerRow) dto.CompanyOwnerResponse {
	return dto.CompanyOwnerResponse{
		EmployeeID: row.ID,
		FirstName:  row.FirstName,
		LastName:   row.LastName,
		Email:      row.Email,
		JobTitle:   row.JobTitle,
	}
}

// Owner godoc
//
//	@Summary		Get a company's account owner
//	@Description	The owner of the account this company belongs to. Ownership is a property of the account, so every company in an account reports the same owner. A company whose account has no owner is 200 with owner:null, not a 404 — that stays reserved for a company that does not exist.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Company id"
//	@Success		200	{object}	dto.CompanyOwnerEnvelope
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/admin/companies/{id}/owner [get]
func (h *AdminCompanyHandler) Owner(c *gin.Context) {
	id, err := adminCompanyParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	ctx := c.Request.Context()
	if _, err := h.q.GetCompanyByID(ctx, id); err != nil {
		apierr.Abort(c, err)
		return
	}

	owner, err := h.q.GetCompanyOwner(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		// An account with no owner is a real, expected state — not a 404 the
		// frontend has to distinguish from "no such company".
		c.JSON(http.StatusOK, dto.CompanyOwnerEnvelope{})
		return
	}
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	resp := toCompanyOwnerResponse(owner)
	c.JSON(http.StatusOK, dto.CompanyOwnerEnvelope{Owner: &resp})
}

// SetOwner godoc
//
//	@Summary		Set a company's account owner
//	@Description	Makes the named employee the owner of the account this company belongs to — the same act as POST /api/v1/admin/accounts/{id}/set-owner, addressed by company. Ownership is a property of the account, so this changes the owner of every company in it. The employee must belong to that account.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int								true	"Company id"
//	@Param			owner	body		dto.SetCompanyOwnerRequest		true	"New owner"
//	@Success		200		{object}	dto.CompanyOwnerEnvelope
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/api/v1/admin/companies/{id}/set-owner [post]
func (h *AdminCompanyHandler) SetOwner(c *gin.Context) {
	id, err := adminCompanyParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	var req dto.SetCompanyOwnerRequest
	if err := bindJSONValidated(c, &req); err != nil {
		apierr.Abort(c, err)
		return
	}

	ctx := c.Request.Context()
	company, err := h.q.GetCompanyByID(ctx, id)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	// The owner must belong to the account being handed over, the same rule
	// /admin/accounts/{id}/set-owner applies. An employee of another client
	// can never own this one.
	belongs, err := h.q.IsEmployeeInAccount(ctx, gen.IsEmployeeInAccountParams{
		EmployeeID: req.EmployeeID,
		AccountID:  &company.AccountID,
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if !belongs {
		apierr.Abort(c, apierr.Validation(map[string]string{
			"employee_id": "must be an employee who belongs to this company's account",
		}))
		return
	}

	if err := h.q.SetAccountOwner(ctx, gen.SetAccountOwnerParams{
		ID:              company.AccountID,
		OwnerEmployeeID: &req.EmployeeID,
	}); err != nil {
		apierr.Abort(c, err)
		return
	}

	owner, err := h.q.GetCompanyOwner(ctx, id)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	resp := toCompanyOwnerResponse(owner)
	c.JSON(http.StatusOK, dto.CompanyOwnerEnvelope{Owner: &resp})
}

// Roles godoc
//
//	@Summary		List another company's roles
//	@Description	The plain /roles route reads company_id off the caller's own JWT, which would show the admin's roles mislabeled as company X's. This reads company X's roles directly, so a provisioning check can answer "did company X get seeded".
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int	true	"Company id"
//	@Param			limit	query		int	false	"Page size"
//	@Param			offset	query		int	false	"Offset"
//	@Success		200		{object}	dto.RolePage
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Router			/api/v1/admin/companies/{id}/roles [get]
func (h *AdminCompanyHandler) Roles(c *gin.Context) {
	id, err := adminCompanyParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	ctx := c.Request.Context()
	if _, err := h.q.GetCompanyByID(ctx, id); err != nil {
		apierr.Abort(c, err)
		return
	}

	p := paginate.Parse(c)
	rows, err := h.q.ListRoles(ctx, gen.ListRolesParams{CompanyID: id, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	total, err := h.q.CountRoles(ctx, id)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	out := make([]dto.RoleResponse, len(rows))
	for i, r := range rows {
		out[i] = toRoleResponse(r)
	}
	c.JSON(http.StatusOK, paginate.NewPage(out, total, p))
}

// WorkOrderStatuses godoc
//
//	@Summary		List another company's work order statuses
//	@Description	The plain /work-order-statuses route reads company_id off the caller's own JWT, which would show the admin's statuses mislabeled as company X's. This reads company X's statuses directly, so a provisioning check can answer "did company X get seeded".
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int	true	"Company id"
//	@Param			limit	query		int	false	"Page size"
//	@Param			offset	query		int	false	"Offset"
//	@Success		200		{object}	dto.WorkOrderStatusPage
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Router			/api/v1/admin/companies/{id}/work-order-statuses [get]
func (h *AdminCompanyHandler) WorkOrderStatuses(c *gin.Context) {
	id, err := adminCompanyParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	ctx := c.Request.Context()
	if _, err := h.q.GetCompanyByID(ctx, id); err != nil {
		apierr.Abort(c, err)
		return
	}

	p := paginate.Parse(c)
	rows, err := h.q.ListWorkOrderStatuses(ctx, gen.ListWorkOrderStatusesParams{CompanyID: id, Limit: int32(p.Limit), Offset: int32(p.Offset)})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	total, err := h.q.CountWorkOrderStatuses(ctx, id)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	out := make([]dto.WorkOrderStatusResponse, len(rows))
	for i, r := range rows {
		out[i] = toWorkOrderStatusResponse(r)
	}
	c.JSON(http.StatusOK, paginate.NewPage(out, total, p))
}
