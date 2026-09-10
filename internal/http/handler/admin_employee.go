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
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/paginate"
)

// AdminEmployeeHandler serves the cross-company employee routes. Everything
// here is deliberately unscoped by the caller's own company_id — that is the
// gap it exists to close — so every route is gated on RequirePlatformAdmin.
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

// membershipChange is one audited addition or removal.
type membershipChange struct {
	company int64
	action  string
}

const (
	membershipGranted = "granted"
	membershipRevoked = "revoked"
)

// membershipDelta is the symmetric difference between what an employee had and
// what a full replace leaves them with — the additions and removals, and only
// those. An id in both sets is membership the request did not touch, and
// recording it would turn the audit into a log of requests rather than of
// changes.
func membershipDelta(before, after []int64) []membershipChange {
	var out []membershipChange
	for _, id := range after {
		if !slices.Contains(before, id) {
			out = append(out, membershipChange{company: id, action: membershipGranted})
		}
	}
	for _, id := range before {
		if !slices.Contains(after, id) {
			out = append(out, membershipChange{company: id, action: membershipRevoked})
		}
	}
	return out
}

// List godoc
//
//	@Summary		List employees across every company
//	@Description	Cross-company employee register. Unlike GET /api/v1/employees this is not scoped to the caller's company: it answers "who exists anywhere", which is what assigning an employee to a second company needs. Pass company_id to narrow it to one tenant's staff — that is membership (employee_companies), not default_company_id, so an employee who belongs to a company that is not their default is still listed.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param		company_id	query		int	false	"Only employees who belong to this company"
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

	// A filter on the existing route rather than a second
	// GET /admin/companies/{id}/employees: that would return the same DTO from a
	// second path, and this one already carries the pagination and the gate.
	companyID, err := queryInt64(c, "company_id")
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	rows, err := h.q.ListAllEmployees(ctx, gen.ListAllEmployeesParams{
		CompanyID: companyID,
		Lim:       int32(p.Limit),
		Off:       int32(p.Offset),
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	// The count takes the same filter, or the page envelope would report a total
	// from a different question than the rows answer.
	total, err := h.q.CountAllEmployees(ctx, companyID)
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
//	@Description	Full overwrite of employee_companies. Requires a platform administrator. company_ids must be non-empty — an employee with no membership cannot log in, so use is_active to deactivate instead. default_company_id must be null or one of company_ids; omitting it clears the employee's stored default_company_id, so send it on every call unless you mean to clear it. Every addition and removal is recorded in membership_audit in the same transaction as the change.
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

	// The delta is what gets audited, so the target's current set is read before
	// the transaction. It used to also be what constrained the caller: a company
	// admin could only add or remove a company they themselves belonged to.
	// That check is gone because the gate replaced it — this namespace now
	// requires a platform administrator, who is entitled to the whole delta by
	// definition, and the old constraint would have blocked exactly the
	// cross-tenant provisioning they exist to do.
	current, err := h.q.ListEmployeeCompanyIDs(ctx, id)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	before := normalizeCompanyIDs(current)

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
		// This route replaces an employee's set of companies; it has no role to
		// grant for any of them, so the membership is created with none. A
		// membership with no role is a legitimate state: the employee is
		// associated with the company but permitted nothing there until someone
		// assigns a role.
		if err := qtx.GrantMembership(ctx, gen.GrantMembershipParams{EmployeeID: id, CompanyID: cid, RoleID: nil}); err != nil {
			apierr.Abort(c, err)
			return
		}
	}
	// Granting a company used to confer administrator access there whenever the
	// employee's role carried is_admin, back when role_id was a single global
	// FK. That is no longer possible: the role now lives on the membership
	// itself, and this route grants none. The audit rows still go in the same
	// transaction as the change, so a recorded grant is one that actually
	// happened.
	now := time.Now().UTC()
	actor := middleware.EmployeeFromContext(ctx)
	for _, cid := range membershipDelta(before, ids) {
		if err := qtx.RecordMembershipChange(ctx, gen.RecordMembershipChangeParams{
			ActorEmployeeID:   &actor,
			SubjectEmployeeID: id,
			CompanyID:         cid.company,
			Action:            cid.action,
			OccurredAt:        now,
		}); err != nil {
			apierr.Abort(c, err)
			return
		}
	}

	if err := qtx.SetEmployeeDefaultCompany(ctx, gen.SetEmployeeDefaultCompanyParams{
		ID:               id,
		DefaultCompanyID: req.DefaultCompanyID,
		UpdatedAt:        now,
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
