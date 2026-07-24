package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"

	"fleet/internal/auth"
)

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
