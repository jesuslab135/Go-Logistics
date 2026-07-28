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
		ctx := context.WithValue(c.Request.Context(), companyContextKey{}, claims.CompanyID)
		ctx = context.WithValue(ctx, employeeContextKey{}, claims.EmployeeID())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
