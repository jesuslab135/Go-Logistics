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
	// IsPlatformAdmin is internal staff acting across tenants, which IsAdmin is
	// not: that one is a tenant's own administrator.
	IsPlatformAdmin bool
	Permissions     map[string]ModulePermissions
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

// Modules is the registry of module names a role's permissions document can key
// on: every module_name Django declared across its viewsets. It exists so
// /me/permissions can report on modules the caller's role says nothing about —
// a role that omits a module denies it, and the client needs to be told that
// rather than left to infer it from an absent key.
//
// Note that three of these are not RequireModule groups in the router: Django
// gated company and roles on IsAdminRole, and tire_approvals on the warehouse
// role check that guarded the approve/reject actions.
var Modules = []string{
	"assets",
	"company",
	"employees",
	"fuel",
	"inspections",
	"inventory",
	"issues",
	"mileage_goals",
	"parts",
	"roles",
	"service",
	"tire_approvals",
	"tires",
	"vendors",
	"warranties",
	"work_orders",
}

// standardActions are the actions every module is reported on, being the ones
// methodAction can produce. Roles may define others (Django's warehouse role
// carries "approve"); those are reported too, but only where a role names them.
var standardActions = []string{"read", "create", "update", "delete"}

// Can ports HasModuleAccess.has_permission.
func (i Identity) Can(module, method string) bool {
	// An unrecognised method yields no action, which allows below treats as
	// "not granted by an explicit action" — matching Django, where an admin or
	// a whole-module grant still passes and everyone else is refused.
	return i.allows(module, methodAction[method])
}

// CanAction checks a named action directly, for permissions that do not
// correspond to an HTTP method (tire_approvals/approve).
func (i Identity) CanAction(module, action string) bool {
	return i.allows(module, action)
}

// allows is the single rule behind both the request gate and the permissions
// report, so what /me/permissions promises cannot drift from what is enforced.
func (i Identity) allows(module, action string) bool {
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
	if action == "" {
		return false
	}
	return perms.Actions[action]
}

// EffectivePermissions expands the caller's role into the answer to "may I do X
// to Y", with the empty-object convention and the admin bypass already applied.
// Clients render from this instead of reimplementing the rules.
func (i Identity) EffectivePermissions() map[string]map[string]bool {
	out := make(map[string]map[string]bool, len(Modules))
	for _, module := range Modules {
		out[module] = i.moduleActions(module)
	}
	// A role may name a module this build's registry does not know yet; report
	// it rather than dropping it silently.
	for module := range i.Permissions {
		if _, ok := out[module]; !ok {
			out[module] = i.moduleActions(module)
		}
	}
	return out
}

func (i Identity) moduleActions(module string) map[string]bool {
	actions := make(map[string]bool, len(standardActions))
	for _, action := range standardActions {
		actions[action] = i.allows(module, action)
	}
	for action := range i.Permissions[module].Actions {
		actions[action] = i.allows(module, action)
	}
	return actions
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

// RequirePlatformAdmin gates the cross-tenant /admin namespace. It is
// deliberately not RequireAdminRole: employee.role_id is a single global FK, so
// a tenant administrator is an administrator in every company they belong to,
// and gating cross-tenant routes on that let any company admin read and rewrite
// another tenant's employees.
func RequirePlatformAdmin() gin.HandlerFunc {
	return gate(func(i Identity, _ *gin.Context) bool { return i.IsPlatformAdmin },
		"platform administrator privileges are required for this action")
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

// RequireAction gates a route on a named action rather than the HTTP method's
// implied one. Approving a tire assignment is a POST, but "create" is not the
// permission it should need — Django guarded it with tire_approvals/approve.
func RequireAction(module, action string) gin.HandlerFunc {
	return gate(func(i Identity, _ *gin.Context) bool { return i.CanAction(module, action) },
		"you do not have permission to perform this action")
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
