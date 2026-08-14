package handler

import (
	"github.com/gin-gonic/gin"

	"fleet/internal/platform/reqbind"
)

// bindJSONValidated binds a request body and reports validation failures as
// per-field errors. The rule lives in reqbind so the generic crud handlers
// apply the same one.
func bindJSONValidated(c *gin.Context, obj any) error {
	return reqbind.JSON(c, obj)
}
