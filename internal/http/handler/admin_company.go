package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/platform/apierr"
)

// AdminCompanyHandler serves the cross-company /admin/companies routes: owner
// read/set here, and (Task 5) the roles and work-order-statuses lists. Every
// route is addressed by explicit company id rather than the caller's own
// membership, which is what the /api/v1/companies namespace already covers.
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
//	@Description	employee.is_account_owner is a single global boolean, not a per-company column: this reads the member of this company who carries it. A company with no owner is 200 with owner:null, not a 404 — that stays reserved for a company that does not exist.
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
		// A company with no owner is a real, expected state — not a 404 the
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
//	@Description	Clears the current owner and flags the named employee in one transaction, so a company never carries two. The employee must already be a member of this company — is_account_owner is a single global boolean, so this cannot make an outsider the owner of a tenant.
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
	if _, err := h.q.GetCompanyByID(ctx, id); err != nil {
		apierr.Abort(c, err)
		return
	}

	now := time.Now().UTC()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	defer tx.Rollback(ctx)
	qtx := h.q.WithTx(tx)

	if err := qtx.ClearCompanyAccountOwner(ctx, gen.ClearCompanyAccountOwnerParams{CompanyID: id, UpdatedAt: now}); err != nil {
		apierr.Abort(c, err)
		return
	}
	owner, err := qtx.SetCompanyAccountOwner(ctx, gen.SetCompanyAccountOwnerParams{ID: req.EmployeeID, CompanyID: id, UpdatedAt: now})
	if errors.Is(err, pgx.ErrNoRows) {
		apierr.Abort(c, apierr.Validation(map[string]string{
			"employee_id": "must be an employee who belongs to this company",
		}))
		return
	}
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		apierr.Abort(c, err)
		return
	}

	resp := dto.CompanyOwnerResponse{
		EmployeeID: owner.ID,
		FirstName:  owner.FirstName,
		LastName:   owner.LastName,
		Email:      owner.Email,
		JobTitle:   owner.JobTitle,
	}
	c.JSON(http.StatusOK, dto.CompanyOwnerEnvelope{Owner: &resp})
}
