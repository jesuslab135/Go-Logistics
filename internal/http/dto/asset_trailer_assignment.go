package dto

import (
	"time"
)

type CreateAssetTrailerAssignmentRequest struct {
	TrailerID      int64      `json:"trailer_id"`
	Position       int32      `json:"position"`
	AssignedDate   time.Time  `json:"assigned_date"`
	UnassignedDate *time.Time `json:"unassigned_date"`
	AssignedByID   *int64     `json:"assigned_by_id"`
	IsActive       *bool      `json:"is_active"`
	Notes          string     `json:"notes"`
}

type UpdateAssetTrailerAssignmentRequest struct {
	TrailerID      int64      `json:"trailer_id"`
	Position       int32      `json:"position"`
	AssignedDate   time.Time  `json:"assigned_date"`
	UnassignedDate *time.Time `json:"unassigned_date"`
	AssignedByID   *int64     `json:"assigned_by_id"`
	IsActive       bool       `json:"is_active"`
	Notes          string     `json:"notes"`
}

type AssetTrailerAssignmentResponse struct {
	ID             int64      `json:"id"`
	AssetID        int64      `json:"asset_id"`
	TrailerID      int64      `json:"trailer_id"`
	Position       int32      `json:"position"`
	AssignedDate   time.Time  `json:"assigned_date"`
	UnassignedDate *time.Time `json:"unassigned_date"`
	AssignedByID   *int64     `json:"assigned_by_id"`
	IsActive       bool       `json:"is_active"`
	Notes          string     `json:"notes"`

	// Denormalized from the trailer asset so a row renders without a second
	// request per assignment.
	TrailerName string `json:"trailer_name"`
}

type AssetTrailerAssignmentPage struct {
	Data    []AssetTrailerAssignmentResponse `json:"data"`
	Total   int64                            `json:"total"`
	Limit   int                              `json:"limit"`
	Offset  int                              `json:"offset"`
	HasNext bool                             `json:"has_next"`
}
