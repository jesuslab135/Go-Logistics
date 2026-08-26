package handler

import (
	"context"
	"testing"

	"fleet/internal/platform/storage"
)

// A file column holds either a storage key (written since uploads became
// visibility-aware) or an absolute URL (written before). Telling them apart is
// what decides whether an object is resolved, re-signed and reclaimed at all, so
// every shape a column can hold is pinned here.
func TestFileKey(t *testing.T) {
	files := storage.NewLocal(t.TempDir(), "/media")

	tests := []struct {
		name   string
		stored string
		want   string
		wantOK bool
	}{
		{"a bare key is already a key", "uploads/private/1/abc-receipt.png", "uploads/private/1/abc-receipt.png", true},
		{"a public key is already a key", "uploads/public/1/abc-photo.png", "uploads/public/1/abc-photo.png", true},
		{"a legacy key with no visibility segment", "uploads/1/abc-photo.png", "uploads/1/abc-photo.png", true},
		{"a URL this backend serves reverses to its key", "/media/uploads/1/old.png", "uploads/1/old.png", true},
		{"an empty column is not a reference", "", "", false},
		{"a foreign host is not ours", "https://example.com/somebody-elses.png", "", false},
		{"a root path this backend does not serve is not ours", "/elsewhere/x.png", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := fileKey(files, tt.stored)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("fileKey(%q) = (%q, %v), want (%q, %v)", tt.stored, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

// A reference the backend cannot resolve must survive into the response rather
// than being blanked: an externally hosted image is still the right thing to
// render, and a broken image is visible where an empty field is not.
func TestFileReadURLPassesThroughWhatItDoesNotOwn(t *testing.T) {
	files := storage.NewLocal(t.TempDir(), "/media")

	const foreign = "https://example.com/somebody-elses.png"
	url, expiresAt := fileReadURL(context.Background(), files, foreign)
	if url != foreign {
		t.Errorf("fileReadURL = %q, want the value passed through unchanged", url)
	}
	if expiresAt != nil {
		t.Errorf("expiry = %v, want nil for a reference we do not sign", expiresAt)
	}
}

// The local backend has no signing mechanism, so it reports no expiry. Pinning
// this keeps a developer from concluding that private objects expire in
// development and being surprised in production, where they do.
func TestFileReadURLLocalNeverExpires(t *testing.T) {
	files := storage.NewLocal(t.TempDir(), "/media")

	url, expiresAt := fileReadURL(context.Background(), files, "uploads/private/1/receipt.png")
	if url != "/media/uploads/private/1/receipt.png" {
		t.Errorf("fileReadURL = %q, want the local URL for the key", url)
	}
	if expiresAt != nil {
		t.Errorf("expiry = %v, want nil: local disk cannot sign", expiresAt)
	}
}

func TestFileReadURLPtrHandlesNull(t *testing.T) {
	files := storage.NewLocal(t.TempDir(), "/media")

	url, expiresAt := fileReadURLPtr(context.Background(), files, nil)
	if url != nil || expiresAt != nil {
		t.Errorf("fileReadURLPtr(nil) = (%v, %v), want (nil, nil)", url, expiresAt)
	}
}
