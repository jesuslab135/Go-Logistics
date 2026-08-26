package dto

import "time"

// UploadResponse is what POST /api/v1/uploads returns. The thumbnail fields are
// empty when none was produced — a non-image, a format the server cannot decode,
// or a generation failure that was not allowed to fail the upload.
//
// Key, not URL, is what callers store on the owning record. A private object's
// URL is signed and expires, so persisting one would store a value that stops
// working; the key is permanent and the server re-signs it on every read.
type UploadResponse struct {
	Key string `json:"key"`
	// URL renders the upload immediately. For a private object it expires at
	// ExpiresAt and must be re-fetched from the owning record, never cached.
	URL string `json:"url"`
	// ExpiresAt is null for a public object, whose URL never expires.
	ExpiresAt *time.Time `json:"expires_at"`
	// Visibility is "public" or "private", determined by the declared purpose.
	Visibility   string `json:"visibility"`
	Size         int64  `json:"size"`
	ContentType  string `json:"content_type"`
	ThumbnailKey string `json:"thumbnail_key"`
	ThumbnailURL string `json:"thumbnail_url"`
}
