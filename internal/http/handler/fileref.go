package handler

import (
	"context"
	"strings"
	"time"

	"fleet/internal/platform/storage"
)

// A file column holds an object *reference*, not a URL.
//
// It used to hold the fully-qualified public URL, which worked only while every
// object was anonymously readable. A private object's URL is signed and expires,
// so persisting one would store a value that stops working — and re-signing it
// on read requires the key, not the dead URL. Columns therefore hold the storage
// key, and the URL is resolved at response time.
//
// Rows written before this change still hold absolute URLs. Rather than rewrite
// them, both forms are accepted on read: a value that looks like a URL is
// reversed through the storage backend, anything else is already a key. The two
// are unambiguous — a key is a relative path and never carries a scheme.

// fileKey normalises a stored file reference to a storage key. It reports false
// for an empty value, or for a URL this storage does not own (an externally
// hosted image), which is a value we must neither resolve nor reclaim.
//
// The backend is asked first rather than the value being sniffed for a scheme:
// the local backend advertises objects at a root-relative base ("/media/..."),
// so "looks absolute" is not what distinguishes a stored URL from a stored key.
// What does is whether the backend recognises it. Anything left over is a key
// unless it carries a location — a key is always a bare relative path.
func fileKey(files storage.Storage, stored string) (string, bool) {
	if stored == "" {
		return "", false
	}
	if files != nil {
		if key, ok := files.KeyFromURL(stored); ok {
			return key, true
		}
	}
	if hasLocation(stored) {
		// A URL with a host or an absolute path that this backend does not serve:
		// someone else's object, which we must not resolve or delete.
		return "", false
	}
	return stored, true
}

// fileReadURL resolves a stored reference to the URL a client should read it at,
// plus the moment that URL expires (nil when it does not). A reference that
// cannot be resolved is returned unchanged: an externally hosted URL is still
// the right thing to render, and a legacy value we no longer own is better shown
// than blanked.
func fileReadURL(ctx context.Context, files storage.Storage, stored string) (string, *time.Time) {
	key, ok := fileKey(files, stored)
	if !ok || files == nil {
		return stored, nil
	}

	url, expiresAt, err := files.ReadURL(ctx, key)
	if err != nil {
		// Signing failed. Returning the stored reference keeps the response
		// well-formed; the image will not render, which is the honest outcome
		// and is visible, unlike a silently blanked field.
		return stored, nil
	}
	if expiresAt.IsZero() {
		return url, nil
	}
	return url, &expiresAt
}

// fileReadURLPtr is fileReadURL for a nullable column.
func fileReadURLPtr(ctx context.Context, files storage.Storage, stored *string) (*string, *time.Time) {
	if stored == nil {
		return nil, nil
	}
	url, expiresAt := fileReadURL(ctx, files, *stored)
	return &url, expiresAt
}

// hasLocation reports whether a value names where it lives — a scheme, or a
// root-absolute path — rather than being a bare storage key.
func hasLocation(v string) bool {
	return strings.HasPrefix(v, "/") ||
		strings.HasPrefix(v, "http://") ||
		strings.HasPrefix(v, "https://")
}
