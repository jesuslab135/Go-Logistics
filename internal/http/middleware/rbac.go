package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"fleet/internal/platform/apierr"
)

// Identity is the authorization subject: who the caller is, resolved from the
// database rather than from the token, so role changes and deactivations apply
// at once. It ports the inputs of api/permissions.py.
type Identity struct {
	EmployeeID     int64
	CompanyID      int64
	IsActive       bool
	IsAccountOwner bool
	IsMember       bool
	HasRole        bool
	IsAdmin        bool
	Permissions    map[string]ModulePermissions
}

// ModulePermissions is one entry of role.permissions. An empty object grants the
// whole module — that is Django's convention, not an accident of decoding.
type ModulePermissions struct {
	All     bool
	Actions map[string]bool
}

// IdentityLoader reads the caller's authorization state. Implemented in the
// handler package over the sqlc queries, so this package stays transport-only.
type IdentityLoader interface {
	LoadIdentity(ctx context.Context, employeeID, companyID int64) (Identity, error)
}

// methodAction ports HasModuleAccess.METHOD_ACTION_MAP.
var methodAction = map[string]string{
	http.MethodGet:     "read",
	http.MethodHead:    "read",
	http.MethodOptions: "read",
	http.MethodPost:    "create",
	http.MethodPut:     "update",
	http.MethodPatch:   "update",
	http.MethodDelete:  "delete",
}

// Can ports HasModuleAccess.has_permission.
func (i Identity) Can(module, method string) bool {
	if !i.IsActive {
		return false
	}
	if i.IsAdmin {
		return true
	}
	if !i.HasRole || module == "" {
		return false
	}

	perms, ok := i.Permissions[module]
	if !ok {
		return false
	}
	if perms.All {
		return true
	}

	// An unrecognised method yields no action, which is refused here unless the
	// checks above already passed — matching Django.
	action, ok := methodAction[method]
	if !ok {
		return false
	}
	return perms.Actions[action]
}

type identityContextKey struct{}

// IdentityOf returns the identity resolved by RequireIdentity, if present.
func IdentityOf(c *gin.Context) (Identity, bool) {
	id, ok := c.Request.Context().Value(identityContextKey{}).(Identity)
	return id, ok
}

// RequireIdentity must run after Auth. It resolves the caller once per request
// so the gates below are pure predicates.
func RequireIdentity(loader IdentityLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := ClaimsOf(c)
		if !ok {
			apierr.Abort(c, apierr.Unauthorized("missing bearer token"))
			return
		}

		identity, err := loader.LoadIdentity(c.Request.Context(), claims.EmployeeID(), claims.CompanyID)
		if err != nil {
			apierr.Abort(c, apierr.Forbidden("no active employee profile is associated with this account").Wrap(err))
			return
		}
		if !identity.IsActive {
			apierr.Abort(c, apierr.Forbidden("this employee account is inactive"))
			return
		}

		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), identityContextKey{}, identity),
		)
		c.Next()
	}
}

// RequireCompanyMember ports IsCompanyMember. Django only checked that the
// employee belonged to *some* company; membership is checked against the token's
// company here, since that is the tenant every scoped query will use.
func RequireCompanyMember() gin.HandlerFunc {
	return gate(func(i Identity, _ *gin.Context) bool { return i.IsMember },
		"no active employee profile is associated with this company")
}

// RequireAdminRole ports IsAdminRole.
func RequireAdminRole() gin.HandlerFunc {
	return gate(func(i Identity, _ *gin.Context) bool { return i.IsAdmin },
		"an administrator role is required for this action")
}

// RequireAccountOwner ports IsAccountOwner.
func RequireAccountOwner() gin.HandlerFunc {
	return gate(func(i Identity, _ *gin.Context) bool { return i.IsAccountOwner },
		"account owner privileges are required for this action")
}

// RequireModule ports HasModuleAccess for a fixed module, replacing the
// per-viewset module_name attribute.
func RequireModule(module string) gin.HandlerFunc {
	return gate(func(i Identity, c *gin.Context) bool { return i.Can(module, c.Request.Method) },
		"you do not have access to this module")
}

func gate(allowed func(Identity, *gin.Context) bool, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, ok := IdentityOf(c)
		if !ok || !allowed(identity, c) {
			apierr.Abort(c, apierr.Forbidden(message))
			return
		}
		c.Next()
	}
}

// DecodePermissions parses a role.permissions jsonb document. Malformed content
// grants nothing rather than failing the request open.
func DecodePermissions(raw []byte) map[string]ModulePermissions {
	out := make(map[string]ModulePermissions)
	if len(raw) == 0 {
		return out
	}

	var modules map[string]json.RawMessage
	if err := json.Unmarshal(raw, &modules); err != nil {
		return out
	}
	for module, body := range modules {
		var actions map[string]bool
		if err := json.Unmarshal(body, &actions); err != nil {
			continue
		}
		out[module] = ModulePermissions{All: len(actions) == 0, Actions: actions}
	}
	return out
}
