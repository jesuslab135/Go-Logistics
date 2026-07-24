package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"fleet/internal/platform/apierr"
)

// Logger emits one structured line per request after it completes.
func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		log.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
			"request_id", RequestIDOf(c),
		)
	}
}

// Recovery converts a panic into a consistent 500 JSON response instead of a
// dropped connection, logging the cause with the request id.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					"error", r,
					"path", c.Request.URL.Path,
					"request_id", RequestIDOf(c),
				)
				apierr.Abort(c, apierr.Internal(fmt.Errorf("panic: %v", r)))
			}
		}()
		c.Next()
	}
}
