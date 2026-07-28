package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateInspectionSubmissionRequest struct {
	FormID             int64            `json:"form_id"`
	AssetID            int64            `json:"asset_id"`
	SubmittedByID      int64            `json:"submitted_by_id"`
	StartedAt          time.Time        `json:"started_at"`
	SubmittedAt        time.Time        `json:"submitted_at"`
	DurationSeconds    *int32           `json:"duration_seconds"`
	StartingLatitude   *decimal.Decimal `json:"starting_latitude"`
	StartingLongitude  *decimal.Decimal `json:"starting_longitude"`
	SubmittedLatitude  *decimal.Decimal `json:"submitted_latitude"`
	SubmittedLongitude *decimal.Decimal `json:"submitted_longitude"`
	Signature          *string          `json:"signature" binding:"omitempty,max=500"`
	Odometer           *decimal.Decimal `json:"odometer"`
	TotalItems         int32            `json:"total_items"`
	FailedItemsCount   int32            `json:"failed_items_count"`
	PassedItemsCount   int32            `json:"passed_items_count"`
	CommentsCount      int32            `json:"comments_count"`
	ImagesCount        int32            `json:"images_count"`
	GeneralNotes       string           `json:"general_notes"`
}

type UpdateInspectionSubmissionRequest struct {
	FormID             int64            `json:"form_id"`
	AssetID            int64            `json:"asset_id"`
	SubmittedByID      int64            `json:"submitted_by_id"`
	StartedAt          time.Time        `json:"started_at"`
	SubmittedAt        time.Time        `json:"submitted_at"`
	DurationSeconds    *int32           `json:"duration_seconds"`
	StartingLatitude   *decimal.Decimal `json:"starting_latitude"`
	StartingLongitude  *decimal.Decimal `json:"starting_longitude"`
	SubmittedLatitude  *decimal.Decimal `json:"submitted_latitude"`
	SubmittedLongitude *decimal.Decimal `json:"submitted_longitude"`
	Signature          *string          `json:"signature" binding:"omitempty,max=500"`
	Odometer           *decimal.Decimal `json:"odometer"`
	TotalItems         int32            `json:"total_items"`
	FailedItemsCount   int32            `json:"failed_items_count"`
	PassedItemsCount   int32            `json:"passed_items_count"`
	CommentsCount      int32            `json:"comments_count"`
	ImagesCount        int32            `json:"images_count"`
	GeneralNotes       string           `json:"general_notes"`
}

type InspectionSubmissionResponse struct {
	ID                 int64            `json:"id"`
	CompanyID          int64            `json:"company_id"`
	FormID             int64            `json:"form_id"`
	AssetID            int64            `json:"asset_id"`
	SubmittedByID      int64            `json:"submitted_by_id"`
	StartedAt          time.Time        `json:"started_at"`
	SubmittedAt        time.Time        `json:"submitted_at"`
	DurationSeconds    *int32           `json:"duration_seconds"`
	StartingLatitude   *decimal.Decimal `json:"starting_latitude"`
	StartingLongitude  *decimal.Decimal `json:"starting_longitude"`
	SubmittedLatitude  *decimal.Decimal `json:"submitted_latitude"`
	SubmittedLongitude *decimal.Decimal `json:"submitted_longitude"`
	Signature          *string          `json:"signature" binding:"omitempty,max=500"`
	Odometer           *decimal.Decimal `json:"odometer"`
	TotalItems         int32            `json:"total_items"`
	FailedItemsCount   int32            `json:"failed_items_count"`
	PassedItemsCount   int32            `json:"passed_items_count"`
	CommentsCount      int32            `json:"comments_count"`
	ImagesCount        int32            `json:"images_count"`
	GeneralNotes       string           `json:"general_notes"`
	CreatedAt          time.Time        `json:"created_at"`
}

type InspectionSubmissionPage struct {
	Data    []InspectionSubmissionResponse `json:"data"`
	Total   int64                          `json:"total"`
	Limit   int                            `json:"limit"`
	Offset  int                            `json:"offset"`
	HasNext bool                           `json:"has_next"`
}
