package middleware

import (
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
		c.Next()
	}
}

// RequireAdmin must run after Auth; it rejects non-admin identities.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsOf(c)
		if !ok || !claims.IsAdmin {
			apierr.Abort(c, apierr.Forbidden("admin privileges required"))
			return
		}
		c.Next()
	}
}
