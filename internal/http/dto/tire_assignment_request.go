package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// A new request is always PENDING and always attributed to the caller. State,
// requester, resolver and timestamps are stamped by the server, so a client
// cannot file a request that is already approved.
type CreateTireAssignmentRequestRequest struct {
	TireID       int64  `json:"tire_id" binding:"required,min=1"`
	VehicleID    int64  `json:"vehicle_id" binding:"required,min=1"`
	PositionCode string `json:"position_code" binding:"required,max=10"`
	Notes        string `json:"notes"`
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

// ApproveTireAssignmentRequestRequest carries the installation readings taken
// at approval time. State, resolver and timestamps are stamped by the server.
type ApproveTireAssignmentRequestRequest struct {
	OdometerAtInstall        int32            `json:"odometer_at_install" binding:"min=0"`
	TreadDepthAtInstall32nds *int32           `json:"tread_depth_at_install_32nds" binding:"omitempty,min=0"`
	PsiAtInstall             *decimal.Decimal `json:"psi_at_install"`
	Notes                    *string          `json:"notes"`
}
