package apierr

import "github.com/gin-gonic/gin"

type body struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Abort maps err to its HTTP form, writes it as JSON, and stops the handler
// chain. The cause is attached to gin's error list for the logging middleware;
// it is never serialized to the client.
func Abort(c *gin.Context, err error) {
	e := Map(err)
	_ = c.Error(err)
	c.AbortWithStatusJSON(e.Status, gin.H{"error": body{Code: e.Code, Message: e.Message, Details: e.Details}})
}
