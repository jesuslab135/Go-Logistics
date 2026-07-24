package dto

import (
	"time"
)

type CreateTireAssignmentRequestRequest struct {
	TireID          int64      `json:"tire_id"`
	VehicleID       int64      `json:"vehicle_id"`
	PositionCode    string     `json:"position_code" binding:"omitempty,max=10"`
	State           string     `json:"state" binding:"omitempty,max=10"`
	RequestedByID   *int64     `json:"requested_by_id"`
	RequestedAt     time.Time  `json:"requested_at"`
	ApprovedByID    *int64     `json:"approved_by_id"`
	ResolvedAt      *time.Time `json:"resolved_at"`
	RejectionReason string     `json:"rejection_reason"`
	Notes           string     `json:"notes"`
}

type UpdateTireAssignmentRequestRequest struct {
	TireID          int64      `json:"tire_id"`
	VehicleID       int64      `json:"vehicle_id"`
	PositionCode    string     `json:"position_code" binding:"omitempty,max=10"`
	State           string     `json:"state" binding:"omitempty,max=10"`
	RequestedByID   *int64     `json:"requested_by_id"`
	RequestedAt     time.Time  `json:"requested_at"`
	ApprovedByID    *int64     `json:"approved_by_id"`
	ResolvedAt      *time.Time `json:"resolved_at"`
	RejectionReason string     `json:"rejection_reason"`
	Notes           string     `json:"notes"`
}

type TireAssignmentRequestResponse struct {
	ID              int64      `json:"id"`
	CompanyID       int64      `json:"company_id"`
	TireID          int64      `json:"tire_id"`
	VehicleID       int64      `json:"vehicle_id"`
	PositionCode    string     `json:"position_code" binding:"omitempty,max=10"`
	State           string     `json:"state" binding:"omitempty,max=10"`
	RequestedByID   *int64     `json:"requested_by_id"`
	RequestedAt     time.Time  `json:"requested_at"`
	ApprovedByID    *int64     `json:"approved_by_id"`
	ResolvedAt      *time.Time `json:"resolved_at"`
	RejectionReason string     `json:"rejection_reason"`
	Notes           string     `json:"notes"`
}

type TireAssignmentRequestPage struct {
	Data    []TireAssignmentRequestResponse `json:"data"`
	Total   int64                           `json:"total"`
	Limit   int                             `json:"limit"`
	Offset  int                             `json:"offset"`
	HasNext bool                            `json:"has_next"`
}
