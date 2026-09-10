package handler

import (
	"errors"
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
)

// AccountEmployeeHandler serves the account-scoped routes that let an account
// owner appoint directors: see who exists in the account and what role each
// person holds in each company, and replace one person's company/role set in
// one call. Everything here is gated on RequireAccountOwner and, unlike
// AdminEmployeeHandler, is scoped to the caller's own account rather than
// crossing tenants.
type AccountEmployeeHandler struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewAccountEmployeeHandler(q *gen.Queries, pool *pgxpool.Pool) *AccountEmployeeHandler {
	return &AccountEmployeeHandler{q: q, pool: pool}
}

// validateGrants refuses two shapes an account owner's request must never
// carry: a company id that cannot possibly name a company, and a company
// named twice, whose second occurrence would silently overwrite the role the
// first one requested with no way for the caller to tell which one won. An
// empty grant list is left alone — it is the legitimate way to revoke every
// membership.
//
// This diverges from admin_employee.go's ReplaceCompanies, which rejects an
// empty company_ids and points the caller at is_active instead: a platform
// admin can still reach that employee through the member-scoped CRUD to flip
// it. An account owner with zero company memberships of their own cannot
// reach that route at all, so refusing empty here would leave them no way to
// revoke anyone's access.
func validateGrants(req dto.ReplaceAccountEmployeeCompaniesRequest) error {
	seen := make(map[int64]bool, len(req.Grants))
	for _, g := range req.Grants {
		if g.CompanyID < 1 {
			return apierr.Validation(map[string]string{"company_id": "must be positive"})
		}
		if seen[g.CompanyID] {
			return apierr.Validation(map[string]string{"company_id": "named more than once"})
		}
		seen[g.CompanyID] = true
	}
	return nil
}

// groupAccountEmployeeMemberships folds the flat, LEFT JOIN'd rows of
// ListAccountEmployeeMemberships into one entry per employee. An employee
// with no membership still produces one row (company_id NULL), which becomes
// an entry with an empty, non-nil Memberships slice rather than being lost.
func groupAccountEmployeeMemberships(rows []gen.ListAccountEmployeeMembershipsRow) []dto.AccountEmployeeResponse {
	out := []dto.AccountEmployeeResponse{}
	var current *dto.AccountEmployeeResponse

	for _, r := range rows {
		if current == nil || current.EmployeeID != r.EmployeeID {
			out = append(out, dto.AccountEmployeeResponse{
				EmployeeID:  r.EmployeeID,
				FirstName:   r.FirstName,
				LastName:    r.LastName,
				Email:       r.Email,
				IsActive:    r.IsActive,
				Memberships: []dto.AccountEmployeeMembership{},
			})
			current = &out[len(out)-1]
		}
		if r.CompanyID == nil {
			continue
		}
		m := dto.AccountEmployeeMembership{
			CompanyID:   *r.CompanyID,
			RoleID:      r.RoleID,
			RoleIsAdmin: r.RoleIsAdmin,
		}
		if r.CompanyName != nil {
			m.CompanyName = *r.CompanyName
		}
		if r.RoleName != nil {
			m.RoleName = *r.RoleName
		}
		current.Memberships = append(current.Memberships, m)
	}
	return out
}

// List godoc
//
//	@Summary		List the account's people and the role each holds per company
//	@Description	Every employee in the caller's account, with their membership set: which companies they belong to and, per company, the role they hold there (if any). Requires account owner. An employee who belongs to no company yet still appears, with an empty memberships list — this is the route that appoints their first one.
//	@Tags			account
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		dto.AccountEmployeeResponse
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		403	{object}	dto.ErrorResponse
//	@Router			/api/v1/account/employees [get]
func (h *AccountEmployeeHandler) List(c *gin.Context) {
	accountID, err := callerAccountID(c)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	rows, err := h.q.ListAccountEmployeeMemberships(c.Request.Context(), accountID)
	if err != nil {
		apierr.Abort(c, err)
		return
	}

	c.JSON(http.StatusOK, groupAccountEmployeeMemberships(rows))
}

// ReplaceCompanies godoc
//
//	@Summary		Replace one person's company memberships and roles
//	@Description	Full overwrite of the target employee's memberships within the caller's own account: every named company gets exactly the role paired with it (including companies the employee already belonged to — this is how a role changes), and every company not named is revoked. RoleID may be omitted for a company, which grants membership with no role: the employee is associated with the company but permitted nothing there until a role is assigned. An empty grants list revokes every membership. Every grant and revoke is recorded in membership_audit in the same transaction as the change. Requires account owner; the target employee and every named company must belong to the caller's own account, or the request 404s rather than confirming they exist elsewhere.
//	@Tags			account
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int										true	"Employee id"
//	@Param			body	body		dto.ReplaceAccountEmployeeCompaniesRequest	true	"Membership set"
//	@Success		200		{object}	dto.AccountEmployeeResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/api/v1/account/employees/{id}/companies [put]
func (h *AccountEmployeeHandler) ReplaceCompanies(c *gin.Context) {
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

	var req dto.ReplaceAccountEmployeeCompaniesRequest
	if err := bindJSONValidated(c, &req); err != nil {
		apierr.Abort(c, err)
		return
	}
	if err := validateGrants(req); err != nil {
		apierr.Abort(c, err)
		return
	}

	ctx := c.Request.Context()

	// The target must belong to the caller's own account. The caller has no
	// legitimate way to know an employee outside it exists, so a mismatch is a
	// 404 — the composite foreign key is the structural backstop, not this.
	target, err := h.q.GetEmployeeByID(ctx, id)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if target.AccountID == nil || *target.AccountID != accountID {
		apierr.Abort(c, apierr.NotFound("employee not found"))
		return
	}

	// The pre-change set is read before the transaction so the audit trail can
	// tell a revoke from a company the request never mentioned.
	currentIDs, err := h.q.ListEmployeeCompanyIDs(ctx, id)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	before := normalizeCompanyIDs(currentIDs)

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	defer tx.Rollback(ctx)
	qtx := h.q.WithTx(tx)

	// after is built as a non-nil slice, even empty, because
	// RevokeMembershipsNotIn compares company_id against it with ANY(...): a nil
	// slice serializes as SQL NULL, under which "NOT (x = ANY(NULL))" is NULL
	// rather than true, and nothing would be revoked. An empty grants list must
	// revoke everything, so the slice has to stay a real, empty array.
	after := make([]int64, 0, len(req.Grants))
	for _, g := range req.Grants {
		inAccount, err := qtx.CompanyInAccount(ctx, gen.CompanyInAccountParams{
			CompanyID: g.CompanyID,
			AccountID: accountID,
		})
		if err != nil {
			apierr.Abort(c, err)
			return
		}
		if !inAccount {
			apierr.Abort(c, apierr.NotFound("company not found"))
			return
		}

		if g.RoleID != nil {
			belongs, err := qtx.RoleBelongsToCompany(ctx, gen.RoleBelongsToCompanyParams{
				RoleID:    *g.RoleID,
				CompanyID: g.CompanyID,
			})
			if err != nil {
				apierr.Abort(c, err)
				return
			}
			if !belongs {
				apierr.Abort(c, apierr.Validation(map[string]string{
					"role_id": "does not belong to the named company",
				}))
				return
			}
		}

		// GrantMembership's ON CONFLICT DO UPDATE SET role_id = EXCLUDED.role_id
		// is exactly what this route wants, unlike admin_employee.go's
		// ReplaceCompanies: every company named here gets the role paired with
		// it, including companies the employee already belonged to, because
		// changing that role is this route's entire purpose.
		if err := qtx.GrantMembership(ctx, gen.GrantMembershipParams{
			EmployeeID: id,
			CompanyID:  g.CompanyID,
			RoleID:     g.RoleID,
		}); err != nil {
			apierr.Abort(c, err)
			return
		}
		after = append(after, g.CompanyID)
	}

	if err := qtx.RevokeMembershipsNotIn(ctx, gen.RevokeMembershipsNotInParams{
		EmployeeID: id,
		CompanyIds: after,
	}); err != nil {
		apierr.Abort(c, err)
		return
	}

	// Every grant is audited, not only the symmetric-difference delta: a grant
	// that re-names a company the employee already belonged to changes the role
	// held there, which is the highest-privilege write an account owner can
	// make when that role carries is_admin, and it must be recorded even though
	// membership itself was already present.
	now := time.Now().UTC()
	actor := middleware.EmployeeFromContext(ctx)
	for _, g := range req.Grants {
		if err := qtx.RecordMembershipChange(ctx, gen.RecordMembershipChangeParams{
			ActorEmployeeID:   &actor,
			SubjectEmployeeID: id,
			CompanyID:         g.CompanyID,
			Action:            membershipGranted,
			OccurredAt:        now,
		}); err != nil {
			apierr.Abort(c, err)
			return
		}
	}
	for _, cid := range before {
		if slices.Contains(after, cid) {
			continue
		}
		if err := qtx.RecordMembershipChange(ctx, gen.RecordMembershipChangeParams{
			ActorEmployeeID:   &actor,
			SubjectEmployeeID: id,
			CompanyID:         cid,
			Action:            membershipRevoked,
			OccurredAt:        now,
		}); err != nil {
			apierr.Abort(c, err)
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		apierr.Abort(c, err)
		return
	}

	rows, err := h.q.ListAccountEmployeeMemberships(ctx, accountID)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	for _, r := range groupAccountEmployeeMemberships(rows) {
		if r.EmployeeID == id {
			c.JSON(http.StatusOK, r)
			return
		}
	}
	// The target was just confirmed to belong to this account, and the listing
	// above is unfiltered, so it not appearing would mean the write and the
	// read disagree about the world.
	apierr.Abort(c, apierr.Internal(errors.New("employee missing from account listing after replace")))
}

// callerAccountID reads the account id off the identity RequireIdentity
// resolved. RequireAccountOwner has already gated the route on IsAccountOwner,
// which is only ever true for an employee who belongs to an account, so a nil
// AccountID here would mean that invariant broke rather than anything a
// request could trigger.
func callerAccountID(c *gin.Context) (int64, error) {
	identity, ok := middleware.IdentityOf(c)
	if !ok || identity.AccountID == nil {
		return 0, apierr.Forbidden("account owner privileges are required for this action")
	}
	return *identity.AccountID, nil
}
