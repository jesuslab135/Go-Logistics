package dto

// UploadResponse is what POST /api/v1/uploads returns. The thumbnail fields are
// empty when none was produced — a non-image, a format the server cannot decode,
// or a generation failure that was not allowed to fail the upload.
type UploadResponse struct {
	Key          string `json:"key"`
	URL          string `json:"url"`
	Size         int64  `json:"size"`
	ContentType  string `json:"content_type"`
	ThumbnailKey string `json:"thumbnail_key"`
	ThumbnailURL string `json:"thumbnail_url"`
}
