package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"fleet/internal/auth"
	"fleet/internal/platform/apierr"
)

const bearerPrefix = "Bearer "

// Auth requires a valid access token. It stores the claims for downstream
// handlers (ClaimsOf) and rejects refresh tokens used as access tokens.
func Auth(tokens *auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, bearerPrefix) {
			apierr.Abort(c, apierr.Unauthorized("missing bearer token"))
			return
		}

		claims, err := tokens.Parse(strings.TrimPrefix(header, bearerPrefix))
		if err != nil {
			apierr.Abort(c, apierr.Unauthorized("invalid or expired token"))
			return
		}
		if claims.Type != auth.TypeAccess {
			apierr.Abort(c, apierr.Unauthorized("access token required"))
			return
		}

		c.Set(claimsKey, claims)
		c.Request = c.Request.WithContext(authContext(c.Request.Context(), claims))
		c.Next()
	}
}

// authContext builds the request context Auth attaches to every downstream
// handler. Extracted so a test can drive the exact code Auth uses rather than
// reimplementing it.
//
// The company id is stored only when the session has one. CompanyFromContext's
// contract is "the company, or 0 if absent" — storing a *int64 there instead of
// an int64 would make its type assertion fail silently and always return 0,
// even for a session that does have a company. A company-less session has no
// tenant to scope to anyway, and every company-scoped route sits behind
// RequireCompanyMember, which refuses such a session before any handler reads
// this.
func authContext(ctx context.Context, claims *auth.Claims) context.Context {
	if claims.CompanyID != nil {
		ctx = context.WithValue(ctx, companyContextKey{}, *claims.CompanyID)
	}
	ctx = context.WithValue(ctx, employeeContextKey{}, claims.EmployeeID())
	return ctx
}
