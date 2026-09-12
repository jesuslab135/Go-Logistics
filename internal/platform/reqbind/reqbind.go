// Package reqbind binds JSON request bodies and reports validation failures as
// per-field errors.
//
// It lives outside the handler package so the generic crud handlers use the
// same rule: a rejected field should say which field and why, whether the route
// was hand-wired or generated.
package reqbind

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	ginbinding "github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"fleet/internal/platform/apierr"
)

// Report failures against the JSON field name the client actually sent, not the
// Go field it bound to.
func init() {
	v, ok := ginbinding.Validator.Engine().(*validator.Validate)
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

// JSON binds the request body into obj. A malformed body is a 400; a
// well-formed body that breaks a validation rule is a 422 carrying one entry
// per offending field, so a client can mark the input that needs fixing
// instead of guessing from a single sentence.
func JSON(c *gin.Context, obj any) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			details := make(map[string]string, len(verrs))
			for _, fe := range verrs {
				details[fe.Field()] = message(fe)
			}
			return apierr.Validation(details)
		}
		return apierr.BadRequest("invalid request body").Wrap(err)
	}
	return nil
}

// Validate runs the same struct validation JSON binding runs, for values that
// did not arrive as a request body, such as the rows of a bulk import. Keys are
// dotted JSON paths ("vehicle.engine_serial"), so a nested field names the
// spreadsheet column it came from. It returns nil when obj is valid.
func Validate(obj any) map[string]string {
	err := ginbinding.Validator.ValidateStruct(obj)
	if err == nil {
		return nil
	}
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return map[string]string{"": err.Error()}
	}
	details := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		path := fe.Namespace()
		if i := strings.IndexByte(path, '.'); i >= 0 {
			path = path[i+1:] // drop the struct's own name
		}
		details[path] = message(fe)
	}
	return details
}

func message(fe validator.FieldError) string {
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
