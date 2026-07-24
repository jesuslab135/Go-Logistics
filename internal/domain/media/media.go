package media

import "time"

// Media — port of api/models/media_model.py:35
// FileType and FileSize were editable=False, derived in save() from the upload.
// UploadedByID references auth_user, not Employee.

type Media struct {
	ID           int64
	CompanyID    int64
	AssetID      int64
	File         string
	Title        string
	Description  string
	FileType     string
	FileSize     int32
	UploadedByID *int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
