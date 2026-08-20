package handler

import (
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/paginate"
)

// AdminEmployeeHandler serves the cross-company employee routes. Everything
// here is deliberately unscoped by the caller's own company_id — that is the
// gap it exists to close — so every route is gated on RequireAdminRole.
type AdminEmployeeHandler struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewAdminEmployeeHandler(q *gen.Queries, pool *pgxpool.Pool) *AdminEmployeeHandler {
	return &AdminEmployeeHandler{q: q, pool: pool}
}

// normalizeCompanyIDs returns the ids sorted and de-duplicated, never nil, so a
// replace is idempotent and the response is stable regardless of request order.
func normalizeCompanyIDs(ids []int64) []int64 {
	out := slices.Clone(ids)
	if out == nil {
		out = []int64{}
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// validateMembershipReplace enforces the two invariants a full replace must
// hold: the employee keeps at least one company, and default_company_id names
// one of the companies the same request is granting.
func validateMembershipReplace(companyIDs []int64, defaultCompanyID *int64) error {
	if len(companyIDs) == 0 {
		return apierr.Validation(map[string]string{
			"company_ids": "at least one company is required; use is_active to deactivate an employee",
		})
	}
	for _, id := range companyIDs {
		if id < 1 {
			return apierr.Validation(map[string]string{"company_ids": "company ids must be positive"})
		}
	}
	if defaultCompanyID != nil && !slices.Contains(companyIDs, *defaultCompanyID) {
		return apierr.Validation(map[string]string{
			"default_company_id": "must be null or one of company_ids",
		})
	}
	return nil
}

// List godoc
//
//	@Summary		List employees across every company
//	@Description	Cross-company employee register. Unlike GET /api/v1/employees this is not scoped to the caller's company: it answers "who exists anywhere", which is what assigning an employee to a second company needs.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param		limit		query		int	false	"Page size"
//	@Param		offset		query		int	false	"Offset"
//	@Success		200			{object}	dto.EmployeePage
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		403			{object}	dto.ErrorResponse
//	@Router			/api/v1/admin/employees [get]
func (h *AdminEmployeeHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	rows, err := h.q.ListAllEmployees(ctx, gen.ListAllEmployeesParams{Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	total, err := h.q.CountAllEmployees(ctx)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	out := make([]dto.EmployeeResponse, len(rows))
	for i, r := range rows {
		out[i] = toEmployeeResponse(r)
	}
	c.JSON(http.StatusOK, paginate.NewPage(out, total, p))
}

// ListCompanies godoc
//
//	@Summary		List an employee's company memberships
//	@Description	The employee's employee_companies rows, sorted and de-duplicated, plus the default_company_id login resolves against.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Employee id"
//	@Success		200	{object}	dto.EmployeeCompaniesResponse
//	@Failure		400	{object}	dto.ErrorResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/api/v1/admin/employees/{id}/companies [get]
func (h *AdminEmployeeHandler) ListCompanies(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		apierr.Abort(c, apierr.BadRequest("invalid id"))
		return
	}

	ctx := c.Request.Context()
	// Read the employee first so a missing one is a 404 rather than an empty
	// membership list that reads as "belongs to nothing".
	row, err := h.q.GetEmployeeByID(ctx, id)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	ids, err := h.q.ListEmployeeCompanyIDs(ctx, id)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.EmployeeCompaniesResponse{
		EmployeeID:       id,
		CompanyIDs:       normalizeCompanyIDs(ids),
		DefaultCompanyID: row.DefaultCompanyID,
	})
}

// ReplaceCompanies godoc
//
//	@Summary		Replace an employee's company memberships
//	@Description	Full overwrite of employee_companies. company_ids must be non-empty — an employee with no membership cannot log in, so use is_active to deactivate instead. default_company_id must be null or one of company_ids.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path		int										true	"Employee id"
//	@Param			companies	body		dto.ReplaceEmployeeCompaniesRequest		true	"Membership set"
//	@Success		200			{object}	dto.EmployeeCompaniesResponse
//	@Failure		400			{object}	dto.ErrorResponse
//	@Failure		401			{object}	dto.ErrorResponse
//	@Failure		403			{object}	dto.ErrorResponse
//	@Failure		404			{object}	dto.ErrorResponse
//	@Failure		422			{object}	dto.ErrorResponse
//	@Router			/api/v1/admin/employees/{id}/companies [put]
func (h *AdminEmployeeHandler) ReplaceCompanies(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		apierr.Abort(c, apierr.BadRequest("invalid id"))
		return
	}

	var req dto.ReplaceEmployeeCompaniesRequest
	if err := bindJSONValidated(c, &req); err != nil {
		apierr.Abort(c, err)
		return
	}

	ids := normalizeCompanyIDs(req.CompanyIDs)
	if err := validateMembershipReplace(ids, req.DefaultCompanyID); err != nil {
		apierr.Abort(c, err)
		return
	}

	ctx := c.Request.Context()
	if _, err := h.q.GetEmployeeByID(ctx, id); err != nil {
		apierr.Abort(c, err)
		return
	}

	// Counting first turns an id that names no company into a 422 naming the
	// offending field, instead of the 409 the foreign key would raise.
	n, err := h.q.CountCompaniesByIDs(ctx, ids)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if n != int64(len(ids)) {
		apierr.Abort(c, apierr.Validation(map[string]string{"company_ids": "one or more companies do not exist"}))
		return
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	defer tx.Rollback(ctx)
	qtx := h.q.WithTx(tx)

	// Delete-then-insert in one transaction: the employee is never observably
	// memberless, and the default is only moved once the set that must contain
	// it is in place.
	if err := qtx.RemoveEmployeeCompaniesNotIn(ctx, gen.RemoveEmployeeCompaniesNotInParams{
		EmployeeID: id,
		CompanyIds: ids,
	}); err != nil {
		apierr.Abort(c, err)
		return
	}
	for _, cid := range ids {
		if err := qtx.AddEmployeeCompany(ctx, gen.AddEmployeeCompanyParams{EmployeeID: id, CompanyID: cid}); err != nil {
			apierr.Abort(c, err)
			return
		}
	}
	if err := qtx.SetEmployeeDefaultCompany(ctx, gen.SetEmployeeDefaultCompanyParams{
		ID:               id,
		DefaultCompanyID: req.DefaultCompanyID,
		UpdatedAt:        time.Now().UTC(),
	}); err != nil {
		apierr.Abort(c, err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		apierr.Abort(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.EmployeeCompaniesResponse{
		EmployeeID:       id,
		CompanyIDs:       ids,
		DefaultCompanyID: req.DefaultCompanyID,
	})
}
