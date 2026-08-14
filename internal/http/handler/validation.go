package handler

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"fleet/internal/platform/apierr"
)

// Report validation failures against the JSON field name a client actually
// sent, not the Go struct field it bound to.
func init() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})
}

// bindJSONValidated binds like ShouldBindJSON but turns validator failures
// into a 422 carrying per-field details, so clients get machine-readable
// field errors instead of one opaque bad_request.
func bindJSONValidated(c *gin.Context, obj any) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			details := make(map[string]string, len(verrs))
			for _, fe := range verrs {
				details[fe.Field()] = validationMessage(fe)
			}
			return apierr.Validation(details)
		}
		return apierr.BadRequest("invalid request body").Wrap(err)
	}
	return nil
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "email":
		return "must be a valid email address"
	case "oneof":
		return fmt.Sprintf("must be one of: %s", strings.ReplaceAll(fe.Param(), " ", ", "))
	default:
		return fmt.Sprintf("failed validation rule %q", fe.Tag())
	}
}
