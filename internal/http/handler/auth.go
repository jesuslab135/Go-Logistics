package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"fleet/internal/auth"
	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/platform/apierr"
)

// ErrInvalidCredentials is returned by a CredentialVerifier when the email or
// password does not match; it maps to a 401.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Identity is the authenticated principal a successful login resolves to.
type Identity struct {
	EmployeeID int64
	CompanyID  int64
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
		apierr.Abort(c, apierr.Unauthorized("invalid credentials"))
		return
	}

	pair, err := h.tokens.Issue(id.EmployeeID, id.CompanyID, id.IsAdmin)
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

	pair, err := h.tokens.Issue(claims.EmployeeID(), claims.CompanyID, claims.IsAdmin)
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
