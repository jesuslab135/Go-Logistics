package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateTireMountLogRequest struct {
	VehicleID       int64            `json:"vehicle_id"`
	PositionCode    string           `json:"position_code" binding:"omitempty,max=10"`
	EventType       string           `json:"event_type" binding:"omitempty,max=10"`
	EventDate       time.Time        `json:"event_date"`
	Odometer        *int32           `json:"odometer"`
	TreadDepth32nds *int32           `json:"tread_depth_32nds"`
	Psi             *decimal.Decimal `json:"psi"`
	PerformedByID   *int64           `json:"performed_by_id"`
	Reason          string           `json:"reason" binding:"omitempty,max=255"`
}

type UpdateTireMountLogRequest struct {
	VehicleID       int64            `json:"vehicle_id"`
	PositionCode    string           `json:"position_code" binding:"omitempty,max=10"`
	EventType       string           `json:"event_type" binding:"omitempty,max=10"`
	EventDate       time.Time        `json:"event_date"`
	Odometer        *int32           `json:"odometer"`
	TreadDepth32nds *int32           `json:"tread_depth_32nds"`
	Psi             *decimal.Decimal `json:"psi"`
	PerformedByID   *int64           `json:"performed_by_id"`
	Reason          string           `json:"reason" binding:"omitempty,max=255"`
}

type TireMountLogResponse struct {
	ID              int64            `json:"id"`
	TireID          int64            `json:"tire_id"`
	VehicleID       int64            `json:"vehicle_id"`
	PositionCode    string           `json:"position_code" binding:"omitempty,max=10"`
	EventType       string           `json:"event_type" binding:"omitempty,max=10"`
	EventDate       time.Time        `json:"event_date"`
	Odometer        *int32           `json:"odometer"`
	TreadDepth32nds *int32           `json:"tread_depth_32nds"`
	Psi             *decimal.Decimal `json:"psi"`
	PerformedByID   *int64           `json:"performed_by_id"`
	Reason          string           `json:"reason" binding:"omitempty,max=255"`
}

type TireMountLogPage struct {
	Data    []TireMountLogResponse `json:"data"`
	Total   int64                  `json:"total"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
	HasNext bool                   `json:"has_next"`
}
