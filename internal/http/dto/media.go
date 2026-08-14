package dto

import (
	"time"
)

// file_type and file_size are intentionally absent: Django derived both from the
// stored file (editable=False), so they are resolved from object storage rather
// than trusted from the client.
// Upload attribution is taken from the authenticated context, not the body.
type CreateMediumRequest struct {
	AssetID     int64  `json:"asset_id"`
	File        string `json:"file" binding:"required,max=500"`
	Title       string `json:"title" binding:"omitempty,max=255"`
	Description string `json:"description"`
}

type UpdateMediumRequest struct {
	AssetID     int64  `json:"asset_id"`
	File        string `json:"file" binding:"required,max=500"`
	Title       string `json:"title" binding:"omitempty,max=255"`
	Description string `json:"description"`
}

type MediumResponse struct {
	ID          int64  `json:"id"`
	CompanyID   int64  `json:"company_id"`
	AssetID     int64  `json:"asset_id"`
	File        string `json:"file" binding:"omitempty,max=500"`
	Title       string `json:"title" binding:"omitempty,max=255"`
	Description string `json:"description"`
	FileType    string `json:"file_type" binding:"omitempty,max=20"`
	FileSize    int32  `json:"file_size"`
	// Thumbnail is empty when none exists: media predating thumbnailing, or a
	// format the server cannot decode.
	Thumbnail    string    `json:"thumbnail"`
	UploadedByID *int64    `json:"uploaded_by_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type MediumPage struct {
	Data    []MediumResponse `json:"data"`
	Total   int64            `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
	HasNext bool             `json:"has_next"`
}
