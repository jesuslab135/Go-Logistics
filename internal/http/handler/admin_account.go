package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/auth"
	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/paginate"
)

// AdminAccountHandler serves the cross-client /admin/accounts routes. Creating a
// client is a platform act: nothing inside a tenant can reach these.
type AdminAccountHandler struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewAdminAccountHandler(q *gen.Queries, pool *pgxpool.Pool) *AdminAccountHandler {
	return &AdminAccountHandler{q: q, pool: pool}
}

// adminAccountParam parses the :id path param shared by every
// AdminAccountHandler method, the same way adminCompanyParam does for
// AdminCompanyHandler.
func adminAccountParam(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, apierr.BadRequest("invalid id")
	}
	return id, nil
}

// minOwnerPasswordLength matches what an operator can reasonably be asked to
// generate. It is not a policy engine; it is a floor that stops an unusable
// credential being provisioned by accident.
const minOwnerPasswordLength = 12

// validateCreateAccount reports every offending field, not just the first: an
// operator provisioning a client should not have to resubmit five times to
// discover the fifth problem.
//
// apierr.Validation's Message is the fixed string "validation failed" — the
// per-field reasons live only in Details, which Error() does not include. The
// message is rebuilt here to name the fields, so a caller inspecting the Go
// error (not just the JSON body) can also tell which ones failed.
func validateCreateAccount(in dto.CreateAccountRequest) error {
	missing := map[string]string{}
	var names []string
	add := func(field, reason string) {
		missing[field] = reason
		names = append(names, field)
	}
	if strings.TrimSpace(in.Name) == "" {
		add("name", "this field is required")
	}
	if strings.TrimSpace(in.OwnerFirstName) == "" {
		add("owner_first_name", "this field is required")
	}
	if strings.TrimSpace(in.OwnerLastName) == "" {
		add("owner_last_name", "this field is required")
	}
	if strings.TrimSpace(in.OwnerEmail) == "" {
		add("owner_email", "this field is required")
	}
	if len(in.OwnerPassword) < minOwnerPasswordLength {
		add("owner_password", "must be at least 12 characters")
	}
	if len(missing) == 0 {
		return nil
	}
	err := apierr.Validation(missing)
	err.Message = fmt.Sprintf("validation failed: %s", strings.Join(names, ", "))
	return err
}

// toAccountResponse renders a row from ListAccountsWithCounts. OwnerEmail is
// deliberately left blank here: the list query does not join to the owner
// employee, so it stays reserved for the moment of creation, where the caller
// already supplied it.
func toAccountResponse(r gen.ListAccountsWithCountsRow) dto.AccountResponse {
	return dto.AccountResponse{
		ID:              r.ID,
		Name:            r.Name,
		OwnerEmployeeID: r.OwnerEmployeeID,
		IsActive:        r.IsActive,
		CreatedAt:       r.CreatedAt,
		CompanyCount:    r.CompanyCount,
		EmployeeCount:   r.EmployeeCount,
	}
}

// Create provisions the account, its owner employee and the owner's password in
// one transaction. The three exist together or not at all: an account with an
// owner who cannot log in is the state this design exists to prevent.
func (h *AdminAccountHandler) Create(c *gin.Context) {
	var in dto.CreateAccountRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		apierr.Abort(c, apierr.BadRequest("invalid request body").Wrap(err))
		return
	}
	if err := validateCreateAccount(in); err != nil {
		apierr.Abort(c, err)
		return
	}

	hash, err := auth.HashPassword(in.OwnerPassword)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	ctx := c.Request.Context()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	defer tx.Rollback(ctx)
	qtx := h.q.WithTx(tx)

	account, err := qtx.CreateAccount(ctx, in.Name)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	// The owner carries account_id, so the account must exist first; the account
	// then learns its owner. That ordering is why owner_employee_id is nullable.
	//
	// CreateAccountOwnerEmployee's RETURNING clause names only id, so sqlc
	// generates a bare int64 here rather than a row struct with an .ID field
	// (the brief's draft assumed the latter).
	ownerID, err := qtx.CreateAccountOwnerEmployee(ctx, gen.CreateAccountOwnerEmployeeParams{
		AccountID:    &account.ID,
		FirstName:    in.OwnerFirstName,
		LastName:     in.OwnerLastName,
		Email:        in.OwnerEmail,
		PasswordHash: hash,
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if err := qtx.SetAccountOwner(ctx, gen.SetAccountOwnerParams{
		ID:              account.ID,
		OwnerEmployeeID: &ownerID,
	}); err != nil {
		apierr.Abort(c, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		apierr.Abort(c, err)
		return
	}

	// The password is deliberately absent from this response.
	c.JSON(http.StatusCreated, dto.AccountResponse{
		ID:              account.ID,
		Name:            account.Name,
		OwnerEmployeeID: &ownerID,
		OwnerEmail:      in.OwnerEmail,
		IsActive:        account.IsActive,
		CreatedAt:       account.CreatedAt,
		EmployeeCount:   1,
	})
}

// List returns every client account, cross-tenant, with the company and
// employee counts a provisioning check needs. Same shape as
// AdminCompanyHandler.Roles/WorkOrderStatuses: paginate.Parse for the window,
// a matching count query for the total.
func (h *AdminAccountHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	p := paginate.Parse(c)

	rows, err := h.q.ListAccountsWithCounts(ctx, gen.ListAccountsWithCountsParams{
		Off: int32(p.Offset),
		Lim: int32(p.Limit),
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	total, err := h.q.CountAccounts(ctx)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	out := make([]dto.AccountResponse, len(rows))
	for i, r := range rows {
		out[i] = toAccountResponse(r)
	}
	c.JSON(http.StatusOK, paginate.NewPage(out, total, p))
}

// SetOwner transfers ownership of an account to an employee who already
// belongs to it. "No such account" (404, via GetAccount) and "that employee is
// not a member of this account" (422) are different failures: the first means
// the id in the URL is wrong, the second means the id in the body is. A caller
// can only act correctly on the difference if they are not collapsed into one
// error.
func (h *AdminAccountHandler) SetOwner(c *gin.Context) {
	id, err := adminAccountParam(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	var req dto.SetAccountOwnerRequest
	if err := bindJSONValidated(c, &req); err != nil {
		apierr.Abort(c, err)
		return
	}

	ctx := c.Request.Context()
	// GetAccount alone decides the 404: a nonexistent account is refused here,
	// before the membership check ever runs.
	if _, err := h.q.GetAccount(ctx, id); err != nil {
		apierr.Abort(c, err)
		return
	}

	belongs, err := h.q.IsEmployeeInAccount(ctx, gen.IsEmployeeInAccountParams{
		EmployeeID: req.EmployeeID,
		AccountID:  &id,
	})
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if !belongs {
		apierr.Abort(c, apierr.Validation(map[string]string{
			"employee_id": "must be an employee who belongs to this account",
		}))
		return
	}

	if err := h.q.SetAccountOwner(ctx, gen.SetAccountOwnerParams{
		ID:              id,
		OwnerEmployeeID: &req.EmployeeID,
	}); err != nil {
		apierr.Abort(c, err)
		return
	}

	updated, err := h.q.GetAccount(ctx, id)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.AccountResponse{
		ID:              updated.ID,
		Name:            updated.Name,
		OwnerEmployeeID: updated.OwnerEmployeeID,
		IsActive:        updated.IsActive,
		CreatedAt:       updated.CreatedAt,
	})
}
