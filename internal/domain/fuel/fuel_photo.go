package fuel

import "time"

// FuelPhoto — port of api/models/fuel_interaction_model.py:22
// UploadedByID references auth_user. FileSize is bigint here but int32 on
// media.Media — inconsistent in the Django source, kept as-is.

type FuelPhoto struct {
	ID           int64
	EntryID      int64
	UploadedByID int64
	File         string
	FileName     string
	FileSize     int64
	MimeType     string
	Description  *string
	UploadedAt   time.Time
	IsPrimary    bool
}
