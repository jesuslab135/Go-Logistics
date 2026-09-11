package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/paginate"
)

// AccountCompanyHandler gives the account owner a view of every company in
// their account, including companies they hold no membership in. Appointing
// directors needs the whole set and each company's roles; /me/permissions
// lists only the caller's own memberships, and /roles only the session's
// company. Gated on RequireAccountOwner and scoped to the caller's account.
type AccountCompanyHandler struct {
	q *gen.Queries
}

func NewAccountCompanyHandler(q *gen.Queries) *AccountCompanyHandler {
	return &AccountCompanyHandler{q: q}
}

// List godoc
//
//	@Summary		List every company in the caller's account
//	@Description	Every company of the caller's account, whether or not the caller holds a membership in it. Requires account owner. Unpaginated, like /account/employees.
//	@Tags			account
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		dto.AccountCompanyResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Router			/api/v1/account/companies [get]
func (h *AccountCompanyHandler) List(c *gin.Context) {
	accountID, err := callerAccountID(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	rows, err := h.q.ListAccountCompanies(c.Request.Context(), accountID)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	out := make([]dto.AccountCompanyResponse, len(rows))
	for i, r := range rows {
		out[i] = dto.AccountCompanyResponse{ID: r.ID, Name: r.Name, Logo: r.Logo}
	}
	c.JSON(http.StatusOK, out)
}

// Roles godoc
//
//	@Summary		List the roles of a company in the caller's account
//	@Description	The roles of any company in the caller's account, to choose the role granted in PUT /api/v1/account/employees/{id}/companies. /api/v1/roles only covers the session's company. Requires account owner. A company outside the caller's account is a 404, never confirming it exists elsewhere.
//	@Tags			account
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
//	@Router			/api/v1/account/companies/{id}/roles [get]
func (h *AccountCompanyHandler) Roles(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		apierr.Abort(c, apierr.BadRequest("invalid id"))
		return
	}

	accountID, err := callerAccountID(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	ctx := c.Request.Context()
	inAccount, err := h.q.CompanyInAccount(ctx, gen.CompanyInAccountParams{CompanyID: id, AccountID: accountID})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if !inAccount {
		apierr.Abort(c, apierr.NotFound("company not found"))
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
