package dto

import (
	"time"
)

type CreateFuelPhotoRequest struct {
	UploadedByID int64     `json:"uploaded_by_id"`
	File         string    `json:"file" binding:"omitempty,max=100"`
	FileName     string    `json:"file_name" binding:"omitempty,max=255"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type" binding:"omitempty,max=100"`
	Description  *string   `json:"description"`
	UploadedAt   time.Time `json:"uploaded_at"`
	IsPrimary    bool      `json:"is_primary"`
}

type UpdateFuelPhotoRequest struct {
	UploadedByID int64     `json:"uploaded_by_id"`
	File         string    `json:"file" binding:"omitempty,max=100"`
	FileName     string    `json:"file_name" binding:"omitempty,max=255"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type" binding:"omitempty,max=100"`
	Description  *string   `json:"description"`
	UploadedAt   time.Time `json:"uploaded_at"`
	IsPrimary    bool      `json:"is_primary"`
}

type FuelPhotoResponse struct {
	ID           int64     `json:"id"`
	EntryID      int64     `json:"entry_id"`
	UploadedByID int64     `json:"uploaded_by_id"`
	File         string    `json:"file" binding:"omitempty,max=100"`
	FileName     string    `json:"file_name" binding:"omitempty,max=255"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type" binding:"omitempty,max=100"`
	Description  *string   `json:"description"`
	UploadedAt   time.Time `json:"uploaded_at"`
	IsPrimary    bool      `json:"is_primary"`
}

type FuelPhotoPage struct {
	Data    []FuelPhotoResponse `json:"data"`
	Total   int64               `json:"total"`
	Limit   int                 `json:"limit"`
	Offset  int                 `json:"offset"`
	HasNext bool                `json:"has_next"`
}
