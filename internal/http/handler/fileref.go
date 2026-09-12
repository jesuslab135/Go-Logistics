package handler

import (
	"context"
	"path"
	"strconv"
	"strings"
	"time"

	"fleet/internal/http/middleware"
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

// uploadKeyRoot is the prefix objectKey writes every upload under.
const uploadKeyRoot = "uploads/"

// fileKeyForCompany resolves a stored reference the way fileKey does, then
// refuses one that belongs to another tenant.
//
// The stored value is client-supplied on fuel photos, asset photos, company
// logos, inspection signatures and submission-item photos — the column takes
// whatever string the caller sends. Resolving it without a tenant check let a
// caller name another company's object and have the server act on it: presign
// it, handing back a working URL to their private receipt, or reclaim it,
// deleting their bytes.
//
// Every key this API writes comes from objectKey, which namespaces it
// uploads/<visibility>/<companyID>/<token>-<name>, so a key under uploads/
// carries its owner in the path and is checkable. Anything else is a
// pre-convention key — necessarily public, since private visibility arrived
// together with the prefix — and is passed through rather than orphaned.
//
// The key is cleaned before it is parsed: fileKey returns a bare relative path
// untouched, so without this "uploads/private/7/../9/x" would present company
// 7's segment while naming company 9's object.
func fileKeyForCompany(files storage.Storage, stored string, companyID int64) (string, bool) {
	key, ok := fileKey(files, stored)
	if !ok {
		return "", false
	}

	clean := path.Clean(key)
	if clean != key {
		// Cleaning changed the path, so the stored value was not the key it
		// appeared to be. Refuse rather than silently acting on the resolved
		// one: nothing this API writes ever needs cleaning.
		return "", false
	}
	if !strings.HasPrefix(clean, uploadKeyRoot) {
		return clean, true // pre-convention key: not ours to adjudicate
	}

	rest := strings.TrimPrefix(clean, uploadKeyRoot)
	head, tail, found := strings.Cut(rest, "/")
	if !found {
		return "", false
	}

	// Two key shapes live under uploads/. The current one carries the
	// visibility that decides which bucket the object is in:
	//
	//	uploads/<public|private>/<companyID>/<token>-<name>
	//
	// Keys written before that split have the tenant directly after the root
	// and are public by definition — deploy/nginx/go-logistics.conf documents
	// exactly this shape as what stored URLs look like:
	//
	//	uploads/<companyID>/<token>-<name>
	//
	// Both name their owner, so both are checked rather than one being waved
	// through as "legacy".
	owner := head
	if head == storage.VisibilityPublic.String() || head == storage.VisibilityPrivate.String() {
		owner, _, found = strings.Cut(tail, "/")
		if !found {
			return "", false
		}
	}

	id, err := strconv.ParseInt(owner, 10, 64)
	if err != nil || id != companyID {
		return "", false
	}
	return clean, true
}

// fileReadURL resolves a stored reference to the URL a client should read it at,
// plus the moment that URL expires (nil when it does not). A reference that
// cannot be resolved is returned unchanged: an externally hosted URL is still
// the right thing to render, and a legacy value we no longer own is better shown
// than blanked.
func fileReadURL(ctx context.Context, files storage.Storage, stored string) (string, *time.Time) {
	// The tenant comes from the request context rather than a parameter: every
	// caller already has it there, and one that forgot to pass it would silently
	// resolve another company's object.
	key, ok := fileKeyForCompany(files, stored, middleware.CompanyFromContext(ctx))
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
