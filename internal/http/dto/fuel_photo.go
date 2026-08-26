package dto

import (
	"time"
)

// Upload attribution is taken from the authenticated context, not the body.
type CreateFuelPhotoRequest struct {
	File        string    `json:"file" binding:"omitempty,max=500"`
	FileName    string    `json:"file_name" binding:"omitempty,max=255"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type" binding:"omitempty,max=100"`
	Description *string   `json:"description"`
	UploadedAt  time.Time `json:"uploaded_at"`
	IsPrimary   bool      `json:"is_primary"`
}

type UpdateFuelPhotoRequest struct {
	File        string    `json:"file" binding:"omitempty,max=500"`
	FileName    string    `json:"file_name" binding:"omitempty,max=255"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type" binding:"omitempty,max=100"`
	Description *string   `json:"description"`
	UploadedAt  time.Time `json:"uploaded_at"`
	IsPrimary   bool      `json:"is_primary"`
}

type FuelPhotoResponse struct {
	ID           int64  `json:"id"`
	EntryID      int64  `json:"entry_id"`
	UploadedByID *int64 `json:"uploaded_by_id"`
	// File is a signed, time-limited URL: a fuel receipt is a business record,
	// not branding, so it is not anonymously readable. Render it, never store it.
	File string `json:"file" binding:"omitempty,max=500"`
	// FileExpiresAt is when File stops resolving. Re-read this record for a fresh
	// URL; null would mean the object is public, which a receipt never is.
	FileExpiresAt *time.Time `json:"file_expires_at"`
	FileName      string     `json:"file_name" binding:"omitempty,max=255"`
	FileSize      int64      `json:"file_size"`
	MimeType      string     `json:"mime_type" binding:"omitempty,max=100"`
	Description   *string    `json:"description"`
	UploadedAt    time.Time  `json:"uploaded_at"`
	IsPrimary     bool       `json:"is_primary"`
}

type FuelPhotoPage struct {
	Data    []FuelPhotoResponse `json:"data"`
	Total   int64               `json:"total"`
	Limit   int                 `json:"limit"`
	Offset  int                 `json:"offset"`
	HasNext bool                `json:"has_next"`
}
