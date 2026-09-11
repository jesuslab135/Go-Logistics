package handler

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"fleet/internal/auth"
	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
)

// ErrInvalidCredentials is returned by a CredentialVerifier when the email or
// password does not match; it maps to a 401.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrNoCompanyMembership is returned when credentials are valid but the employee
// belongs to no company. Issuing a token here would scope the session to a
// company that does not exist, so the login is refused instead.
var ErrNoCompanyMembership = errors.New("no company membership")

// ErrInactiveEverywhere is returned when credentials are valid but every
// company membership the employee holds is deactivated.
var ErrInactiveEverywhere = errors.New("inactive in every company")

// Identity is the authenticated principal a successful login resolves to.
// CompanyID and AccountID are pointers: a company-less session (an account
// owner who has not yet created a company) has neither a tenant to scope to
// nor, in CompanyID's case, one to omit as a false zero.
type Identity struct {
	EmployeeID int64
	CompanyID  *int64
	AccountID  *int64
	IsAdmin    bool
}

// CredentialVerifier checks an email/password pair. It is intentionally a seam:
// the current schema stores no passwords (auth_user was not ported), so the
// backing store is a pending decision — inject an implementation to enable login.
type CredentialVerifier interface {
	Verify(ctx context.Context, email, password string) (Identity, error)
}

type AuthHandler struct {
	tokens   *auth.TokenService
	verifier CredentialVerifier
	q        *gen.Queries
}

func NewAuthHandler(tokens *auth.TokenService, verifier CredentialVerifier, q *gen.Queries) *AuthHandler {
	return &AuthHandler{tokens: tokens, verifier: verifier, q: q}
}

// Login godoc
//
//	@Summary	Log in with email and password
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		credentials	body		dto.LoginRequest	true	"Credentials"
//	@Success	200			{object}	auth.TokenPair
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Router		/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	if h.verifier == nil {
		apierr.Abort(c, apierr.New(http.StatusNotImplemented, "not_implemented", "credential verification is not configured"))
		return
	}

	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.Abort(c, apierr.BadRequest("invalid request body").Wrap(err))
		return
	}

	id, err := h.verifier.Verify(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInactiveEverywhere) {
			apierr.Abort(c, apierr.New(http.StatusForbidden, "inactive",
				"this employee is deactivated in every company they belong to"))
			return
		}
		if errors.Is(err, ErrNoCompanyMembership) {
			apierr.Abort(c, apierr.New(http.StatusForbidden, "no_company_membership",
				"this account is not a member of any company"))
			return
		}
		apierr.Abort(c, apierr.Unauthorized("invalid credentials"))
		return
	}

	pair, err := h.tokens.Issue(id.EmployeeID, id.CompanyID, id.AccountID, id.IsAdmin)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, pair)
}

// Refresh godoc
//
//	@Summary	Exchange a refresh token for a new token pair
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		refresh	body		dto.RefreshRequest	true	"Refresh token"
//	@Success	200		{object}	auth.TokenPair
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Failure	403		{object}	dto.ErrorResponse
//	@Router		/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.Abort(c, apierr.BadRequest("invalid request body").Wrap(err))
		return
	}

	claims, err := h.tokens.ParseRefresh(req.RefreshToken)
	if err != nil {
		apierr.Abort(c, apierr.Unauthorized("invalid refresh token"))
		return
	}
	// Tokens minted before BR-05 carry no jti and so can never be revoked;
	// refuse them rather than honour an unrevocable credential.
	if claims.ID == "" {
		apierr.Abort(c, apierr.Unauthorized("invalid refresh token"))
		return
	}
	revoked, err := h.q.IsRefreshTokenRevoked(c.Request.Context(), claims.ID)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	if revoked {
		apierr.Abort(c, apierr.Unauthorized("invalid refresh token"))
		return
	}

	// Login and switch-company both prove employee_companies membership before
	// minting a token; refresh re-issues company_id off the old token, so
	// without this a revoked membership stays live for a whole refresh TTL.
	// A company-less token (claims.CompanyID == nil) has no membership to
	// re-prove — it is simply re-issued company-less.
	if claims.CompanyID != nil {
		companies, err := h.q.ListActiveEmployeeCompanyIDs(c.Request.Context(), claims.EmployeeID())
		if err != nil {
			apierr.Abort(c, err)
			return
		}
		if !slices.Contains(companies, *claims.CompanyID) {
			apierr.Abort(c, apierr.New(http.StatusForbidden, "no_company_membership",
				"this account is no longer a member of the company this token was issued for"))
			return
		}
	}

	pair, err := h.tokens.Issue(claims.EmployeeID(), claims.CompanyID, claims.AccountID, claims.IsAdmin)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, pair)
}

// SwitchCompany godoc
//
//	@Summary		Switch the active company
//	@Description	Re-issues the token pair scoped to another company the caller belongs to.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			company	body		dto.SwitchCompanyRequest	true	"Target company"
//	@Success		200		{object}	auth.TokenPair
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		403		{object}	dto.ErrorResponse
//	@Router			/auth/switch-company [post]
func (h *AuthHandler) SwitchCompany(c *gin.Context) {
	var req dto.SwitchCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.Abort(c, apierr.BadRequest("invalid request body").Wrap(err))
		return
	}

	claims, ok := middleware.ClaimsOf(c)
	if !ok {
		apierr.Abort(c, apierr.Unauthorized("authentication required"))
		return
	}

	// Membership is re-read from employee_companies rather than trusted from
	// the presented token, so a stale claim cannot widen access.
	row, err := h.q.GetEmployeeIdentity(c.Request.Context(), gen.GetEmployeeIdentityParams{
		CompanyID: &req.CompanyID,
		ID:        claims.EmployeeID(),
	})
	if err != nil || !row.IsActive || !row.IsMember {
		apierr.Abort(c, apierr.Forbidden("you are not a member of this company"))
		return
	}

	// Both tokens are re-scoped: is_admin is resolved against the target
	// company's role, not carried over from the previous tenant. account_id is
	// re-read alongside membership rather than trusted from the old token.
	pair, err := h.tokens.Issue(claims.EmployeeID(), &req.CompanyID, row.AccountID, row.IsAdmin)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, pair)
}

// Logout godoc
//
//	@Summary		Revoke a refresh token
//	@Description	Adds the token's jti to the denylist until its natural expiry.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			refresh	body	dto.LogoutRequest	true	"Refresh token"
//	@Success		204		"revoked"
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Router			/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.Abort(c, apierr.BadRequest("invalid request body").Wrap(err))
		return
	}

	claims, err := h.tokens.ParseRefresh(req.RefreshToken)
	if err != nil || claims.ID == "" {
		apierr.Abort(c, apierr.Unauthorized("invalid refresh token"))
		return
	}

	ctx := c.Request.Context()
	// The jti is bound to its own employee, so revoking only ever affects the
	// presented token — a caller cannot revoke someone else's session.
	if err := h.q.RevokeRefreshToken(ctx, gen.RevokeRefreshTokenParams{
		Jti:        claims.ID,
		EmployeeID: claims.EmployeeID(),
		ExpiresAt:  claims.ExpiresAt.Time,
	}); err != nil {
		apierr.Abort(c, err)
		return
	}
	// Opportunistic pruning; a failure here must not fail the logout.
	_ = h.q.DeleteExpiredRevokedTokens(ctx)

	c.Status(http.StatusNoContent)
}
