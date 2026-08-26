package storage

import "testing"

// Visibility is derived from the key, so the key prefix is load-bearing: it is
// what routes an object to the private bucket on write, read, delete and stat.
// A change here silently republishes every affected object, so it is pinned.
func TestVisibilityOf(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want Visibility
	}{
		{"a private upload", "uploads/private/1/abc-receipt.png", VisibilityPrivate},
		{"a public upload", "uploads/public/1/abc-photo.png", VisibilityPublic},
		{"a key written before the split is public, as it already was", "uploads/1/abc-photo.png", VisibilityPublic},
		{"a leading slash does not change the answer", "/uploads/private/1/abc.png", VisibilityPrivate},
		{"a path that merely mentions private is not private", "uploads/public/1/private-notes.png", VisibilityPublic},
		{"an empty key is public, and invalid everywhere else", "", VisibilityPublic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VisibilityOf(tt.key); got != tt.want {
				t.Errorf("VisibilityOf(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

// KeyPrefix and VisibilityOf must agree, or an object would be written under a
// prefix that later reads classify differently — the one failure mode that would
// serve a private object from the public bucket.
func TestKeyPrefixRoundTrips(t *testing.T) {
	for _, vis := range []Visibility{VisibilityPublic, VisibilityPrivate} {
		key := vis.KeyPrefix() + "1/abc-file.png"
		if got := VisibilityOf(key); got != vis {
			t.Errorf("VisibilityOf(%q) = %v, want %v", key, got, vis)
		}
	}
}
