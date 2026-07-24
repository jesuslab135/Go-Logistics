package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"fleet/internal/auth"
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
}

func NewAuthHandler(tokens *auth.TokenService, verifier CredentialVerifier) *AuthHandler {
	return &AuthHandler{tokens: tokens, verifier: verifier}
}

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

	pair, err := h.tokens.Issue(claims.EmployeeID(), claims.CompanyID, claims.IsAdmin)
	if err != nil {
		apierr.Abort(c, err)
		return
	}
	c.JSON(http.StatusOK, pair)
}
