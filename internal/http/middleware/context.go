package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"

	"fleet/internal/auth"
)

// companyContextKey carries the authenticated tenant id on the request context
// so repository stores can scope every query without depending on gin.
type companyContextKey struct{}

// employeeContextKey carries the authenticated employee id, for the few stores
// that must act on behalf of the caller rather than only scope by tenant.
type employeeContextKey struct{}

// CompanyFromContext returns the tenant company id set by Auth, or 0 if absent.
func CompanyFromContext(ctx context.Context) int64 {
	id, _ := ctx.Value(companyContextKey{}).(int64)
	return id
}

// EmployeeFromContext returns the authenticated employee id set by Auth, or 0.
func EmployeeFromContext(ctx context.Context) int64 {
	id, _ := ctx.Value(employeeContextKey{}).(int64)
	return id
}

const (
	RequestIDHeader = "X-Request-ID"
	requestIDKey    = "request_id"
	claimsKey       = "auth_claims"
)

// RequestID assigns each request a stable id (honoring an inbound header) and
// echoes it back, so logs and clients can correlate a single request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = newID()
		}
		c.Set(requestIDKey, id)
		c.Header(RequestIDHeader, id)
		c.Next()
	}
}

func RequestIDOf(c *gin.Context) string {
	if v, ok := c.Get(requestIDKey); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// ClaimsOf returns the authenticated claims set by Auth, if present.
func ClaimsOf(c *gin.Context) (*auth.Claims, bool) {
	v, ok := c.Get(claimsKey)
	if !ok {
		return nil, false
	}
	claims, ok := v.(*auth.Claims)
	return claims, ok
}

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
