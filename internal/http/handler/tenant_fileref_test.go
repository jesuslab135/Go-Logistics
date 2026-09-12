package handler

import (
	"context"
	"testing"

	"fleet/internal/http/middleware"
	"fleet/internal/platform/storage"
)

// A stored file reference is client-supplied on fuel photos, asset photos,
// company logos, inspection signatures and submission-item photos — the column
// takes any string the caller sends. Every key this API writes comes from
// objectKey, which namespaces it uploads/<visibility>/<companyID>/..., so a key
// under uploads/ carries its owner in the path and must match the caller.
//
// Without the check, a client stores another tenant's key and the server either
// presigns it (handing back a working URL to their private receipt) or reclaims
// it (deleting their object).
func TestFileKeyForCompanyRejectsForeignKeys(t *testing.T) {
	const me = int64(7)
	for _, tt := range []struct {
		name    string
		stored  string
		wantOK  bool
		wantKey string
	}{
		{"own public key", "uploads/public/7/ab12-photo.jpg", true, "uploads/public/7/ab12-photo.jpg"},
		{"own private key", "uploads/private/7/ab12-receipt.pdf", true, "uploads/private/7/ab12-receipt.pdf"},
		{"another tenant's private key", "uploads/private/9/ab12-receipt.pdf", false, ""},
		{"another tenant's public key", "uploads/public/9/ab12-photo.jpg", false, ""},
		{"uploads path with no company segment", "uploads/private/secret.pdf", false, ""},
		{"uploads path with a non-numeric company", "uploads/private/x/secret.pdf", false, ""},
		{"traversal out of my own prefix", "uploads/private/7/../9/secret.pdf", false, ""},
		{"traversal above uploads", "uploads/private/7/../../../etc/passwd", false, ""},
		{"company id that merely starts the same", "uploads/private/70/secret.pdf", false, ""},
		// Keys written before the public/private split put the tenant straight
		// after the root. deploy/nginx/go-logistics.conf documents this shape as
		// what stored URLs look like, so production holds them: they are checked
		// on the same rule, not waved through.
		{"pre-split key, own company", "uploads/7/ab12-photo.jpg", true, "uploads/7/ab12-photo.jpg"},
		{"pre-split key, another company", "uploads/9/ab12-photo.jpg", false, ""},
		{"pre-split key with no tenant at all", "uploads/photo.jpg", false, ""},
		{"key outside uploads entirely", "media/2026/07/photo.jpg", true, "media/2026/07/photo.jpg"},
		{"empty", "", false, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			key, ok := fileKeyForCompany(nil, tt.stored, me)
			if ok != tt.wantOK {
				t.Fatalf("fileKeyForCompany(%q, %d) ok = %v, want %v", tt.stored, me, ok, tt.wantOK)
			}
			if ok && key != tt.wantKey {
				t.Errorf("key = %q, want %q", key, tt.wantKey)
			}
		})
	}
}

// The caller's own uploads must keep working: objectKey builds exactly this
// shape, so a false negative here would break every upload round-trip.
func TestFileKeyForCompanyAcceptsWhatObjectKeyProduces(t *testing.T) {
	const company = int64(42)
	for _, vis := range []storage.Visibility{storage.VisibilityPublic, storage.VisibilityPrivate} {
		key := objectKey(vis, company, "informe final.pdf")
		got, ok := fileKeyForCompany(nil, key, company)
		if !ok {
			t.Errorf("objectKey produced %q, which fileKeyForCompany refuses for its own company", key)
			continue
		}
		if got != key {
			t.Errorf("key round-trip changed %q into %q", key, got)
		}
	}
}

// tenantCtx is the context a request would carry after Auth: the file helpers
// read the tenant from it, so a test that omits one is testing an anonymous
// caller rather than the case it means to.
func tenantCtx(companyID int64) context.Context {
	return middleware.ContextWithCompany(context.Background(), companyID)
}
