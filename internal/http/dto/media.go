package dto

import (
	"time"
)

type CreateMediumRequest struct {
	AssetID      int64  `json:"asset_id"`
	File         string `json:"file" binding:"omitempty,max=100"`
	Title        string `json:"title" binding:"omitempty,max=255"`
	Description  string `json:"description"`
	FileType     string `json:"file_type" binding:"omitempty,max=20"`
	FileSize     int32  `json:"file_size"`
	UploadedByID *int64 `json:"uploaded_by_id"`
}

type UpdateMediumRequest struct {
	AssetID      int64  `json:"asset_id"`
	File         string `json:"file" binding:"omitempty,max=100"`
	Title        string `json:"title" binding:"omitempty,max=255"`
	Description  string `json:"description"`
	FileType     string `json:"file_type" binding:"omitempty,max=20"`
	FileSize     int32  `json:"file_size"`
	UploadedByID *int64 `json:"uploaded_by_id"`
}

type MediumResponse struct {
	ID           int64     `json:"id"`
	CompanyID    int64     `json:"company_id"`
	AssetID      int64     `json:"asset_id"`
	File         string    `json:"file" binding:"omitempty,max=100"`
	Title        string    `json:"title" binding:"omitempty,max=255"`
	Description  string    `json:"description"`
	FileType     string    `json:"file_type" binding:"omitempty,max=20"`
	FileSize     int32     `json:"file_size"`
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
