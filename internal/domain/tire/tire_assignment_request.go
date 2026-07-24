package tire

import "time"

// TireAssignmentRequest — port of api/models/tire_model.py:477
// Warehouse approval workflow before a tire may be mounted.

type TireAssignmentRequest struct {
	ID              int64
	CompanyID       int64
	TireID          int64
	VehicleID       int64
	PositionCode    string
	State           string
	RequestedByID   *int64
	RequestedAt     time.Time
	ApprovedByID    *int64
	ResolvedAt      *time.Time
	RejectionReason string
	Notes           string
}
