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
