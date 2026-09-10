package middleware

import (
	"net/http"
	"testing"
)

func TestDecodePermissions(t *testing.T) {
	perms := DecodePermissions([]byte(`{"assets":{},"fuel":{"read":true,"create":false},"bad":42}`))

	if got, ok := perms["assets"]; !ok || !got.All {
		t.Fatalf("empty object must grant the whole module, got %+v", got)
	}
	if got := perms["fuel"]; got.All || !got.Actions["read"] || got.Actions["create"] {
		t.Fatalf("per-action decoding wrong, got %+v", got)
	}
	if _, ok := perms["bad"]; ok {
		t.Fatal("undecodable module entry must be dropped, not granted")
	}
}

func TestDecodePermissionsMalformed(t *testing.T) {
	for _, raw := range [][]byte{nil, []byte(``), []byte(`not json`), []byte(`[]`)} {
		if got := DecodePermissions(raw); len(got) != 0 {
			t.Fatalf("DecodePermissions(%q) = %+v, want empty", raw, got)
		}
	}
}

func TestIdentityCan(t *testing.T) {
	base := Identity{IsActive: true, HasRole: true}

	tests := []struct {
		name     string
		identity Identity
		module   string
		method   string
		want     bool
	}{
		{"inactive employee is denied", Identity{IsAdmin: true}, "assets", http.MethodGet, false},
		{"admin bypasses modules", Identity{IsActive: true, IsAdmin: true}, "assets", http.MethodDelete, true},
		{"no role denies", Identity{IsActive: true}, "assets", http.MethodGet, false},
		{
			"missing module denies",
			withPerms(base, `{"fuel":{}}`), "assets", http.MethodGet, false,
		},
		{
			"empty module object grants every action",
			withPerms(base, `{"assets":{}}`), "assets", http.MethodDelete, true,
		},
		{
			"granted action allowed",
			withPerms(base, `{"assets":{"read":true}}`), "assets", http.MethodGet, true,
		},
		{
			"ungranted action denied",
			withPerms(base, `{"assets":{"read":true}}`), "assets", http.MethodPost, false,
		},
		{
			"explicitly false action denied",
			withPerms(base, `{"assets":{"delete":false}}`), "assets", http.MethodDelete, false,
		},
		{
			"patch maps to update",
			withPerms(base, `{"assets":{"update":true}}`), "assets", http.MethodPatch, true,
		},
		{
			"unmapped method denied",
			withPerms(base, `{"assets":{"read":true}}`), "assets", http.MethodConnect, false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.identity.Can(tt.module, tt.method); got != tt.want {
				t.Fatalf("Can(%q, %s) = %v, want %v", tt.module, tt.method, got, tt.want)
			}
		})
	}
}

func withPerms(base Identity, raw string) Identity {
	base.Permissions = DecodePermissions([]byte(raw))
	return base
}

// What /me/permissions reports must be what the gates enforce, so the report is
// derived from the same rule rather than restating it.
func TestEffectivePermissionsMatchesCan(t *testing.T) {
	identities := map[string]Identity{
		"admin":        {IsActive: true, IsAdmin: true},
		"inactive":     {IsAdmin: true},
		"roleless":     {IsActive: true},
		"whole module": withPerms(Identity{IsActive: true, HasRole: true}, `{"assets":{}}`),
		"per action": withPerms(Identity{IsActive: true, HasRole: true},
			`{"assets":{"read":true,"delete":false},"fuel":{"create":true}}`),
	}

	methodOf := map[string]string{
		"read": http.MethodGet, "create": http.MethodPost,
		"update": http.MethodPut, "delete": http.MethodDelete,
	}

	for name, identity := range identities {
		t.Run(name, func(t *testing.T) {
			effective := identity.EffectivePermissions()

			for _, module := range Modules {
				actions, ok := effective[module]
				if !ok {
					t.Fatalf("module %q missing from the report", module)
				}
				for action, method := range methodOf {
					if got, want := actions[action], identity.Can(module, method); got != want {
						t.Errorf("%s.%s = %v, but Can(%s) = %v", module, action, got, method, want)
					}
				}
			}
		})
	}
}

// A role may grant an action no HTTP method maps to (Django's warehouse role
// carries "approve"), and may name a module this build does not know about.
// Neither may be dropped from the report.
func TestEffectivePermissionsKeepsCustomActionsAndModules(t *testing.T) {
	identity := withPerms(Identity{IsActive: true, HasRole: true},
		`{"tire_approvals":{"approve":true},"future_module":{"read":true}}`)

	effective := identity.EffectivePermissions()

	if !effective["tire_approvals"]["approve"] {
		t.Error("custom action \"approve\" was dropped from the report")
	}
	if _, ok := effective["future_module"]; !ok {
		t.Error("module outside the registry was dropped from the report")
	}
	if !effective["future_module"]["read"] {
		t.Error("unregistered module reported read = false, want true")
	}
}

// An admin's report must grant custom actions too, not just the four standard
// ones — the gate would let them through.
func TestEffectivePermissionsAdminGrantsCustomAction(t *testing.T) {
	identity := withPerms(Identity{IsActive: true, IsAdmin: true, HasRole: true},
		`{"tire_approvals":{"approve":false}}`)

	if !identity.EffectivePermissions()["tire_approvals"]["approve"] {
		t.Error("admin reported approve = false, but the gate would allow it")
	}
}

// RequireAccountOwner used to read a global boolean whose only meaning was "may
// create companies". It now means the caller owns the account they belong to,
// so the predicate must be false for a member of an account they do not own.
func TestAccountOwnerPredicate(t *testing.T) {
	owner := Identity{IsActive: true, IsAccountOwner: true}
	if !owner.IsAccountOwner {
		t.Fatal("an account owner should satisfy the predicate")
	}
	member := Identity{IsActive: true, IsAccountOwner: false}
	if member.IsAccountOwner {
		t.Fatal("a non-owner must not satisfy the predicate")
	}
}

// A company-less session has no tenant. It must not be reported as a member of
// anything, because every scoped query trusts that flag.
func TestCompanylessIdentityIsNotAMember(t *testing.T) {
	id := Identity{IsActive: true, AccountID: i64ptr(3), IsMember: false}
	if id.IsMember {
		t.Fatal("a company-less identity must not be a member")
	}
	if id.AccountID == nil {
		t.Fatal("a company-less identity still belongs to an account")
	}
}

// A membership with no role grants nothing. Forgetting to assign one must fail
// closed: the person is a member, and every module gate still refuses.
func TestMemberWithoutRoleIsRefusedEveryModule(t *testing.T) {
	id := Identity{IsActive: true, IsMember: true, HasRole: false, IsAdmin: false}
	for _, module := range Modules {
		if id.Can(module, "GET") {
			t.Fatalf("module %q allowed a read for a member holding no role", module)
		}
		if id.Can(module, "POST") {
			t.Fatalf("module %q allowed a write for a member holding no role", module)
		}
	}
}

// Ownership is a backstop: with roles on memberships it becomes possible to
// remove your own admin role from a company you own, and nobody in the account
// could then repair it.
func TestAccountOwnerIsAdminWithoutARole(t *testing.T) {
	id := Identity{IsActive: true, IsMember: true, HasRole: false, IsAdmin: true}
	if !id.Can("assets", "DELETE") {
		t.Fatal("an account owner must retain admin without holding a role")
	}
}

func i64ptr(v int64) *int64 { return &v }

func TestIdentityCanAction(t *testing.T) {
	warehouse := Identity{
		IsActive: true,
		HasRole:  true,
		Permissions: map[string]ModulePermissions{
			"tire_approvals": {Actions: map[string]bool{"approve": true}},
		},
	}
	reader := Identity{
		IsActive: true,
		HasRole:  true,
		Permissions: map[string]ModulePermissions{
			"tire_approvals": {Actions: map[string]bool{"read": true}},
		},
	}
	admin := Identity{IsActive: true, HasRole: true, IsAdmin: true}
	wholeModule := Identity{
		IsActive:    true,
		HasRole:     true,
		Permissions: map[string]ModulePermissions{"tire_approvals": {All: true}},
	}

	tests := []struct {
		name     string
		identity Identity
		want     bool
	}{
		{"explicit approve grant", warehouse, true},
		{"read only cannot approve", reader, false},
		{"admin bypasses", admin, true},
		{"whole-module grant covers approve", wholeModule, true},
		{"inactive is refused", Identity{IsActive: false, IsAdmin: true}, false},
		{"no role is refused", Identity{IsActive: true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.identity.CanAction("tire_approvals", "approve"); got != tt.want {
				t.Errorf("CanAction = %v, want %v", got, tt.want)
			}
		})
	}
}
