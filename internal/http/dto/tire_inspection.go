package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateTireInspectionRequest struct {
	InspectionDate  time.Time        `json:"inspection_date"`
	Odometer        *int32           `json:"odometer"`
	TreadDepth32nds int32            `json:"tread_depth_32nds"`
	Psi             *decimal.Decimal `json:"psi"`
	MeasuredByID    *int64           `json:"measured_by_id"`
	Notes           string           `json:"notes"`
}

type UpdateTireInspectionRequest struct {
	InspectionDate  time.Time        `json:"inspection_date"`
	Odometer        *int32           `json:"odometer"`
	TreadDepth32nds int32            `json:"tread_depth_32nds"`
	Psi             *decimal.Decimal `json:"psi"`
	MeasuredByID    *int64           `json:"measured_by_id"`
	Notes           string           `json:"notes"`
}

type TireInspectionResponse struct {
	ID              int64            `json:"id"`
	TireID          int64            `json:"tire_id"`
	InspectionDate  time.Time        `json:"inspection_date"`
	Odometer        *int32           `json:"odometer"`
	TreadDepth32nds int32            `json:"tread_depth_32nds"`
	Psi             *decimal.Decimal `json:"psi"`
	MeasuredByID    *int64           `json:"measured_by_id"`
	Notes           string           `json:"notes"`
}

type TireInspectionPage struct {
	Data    []TireInspectionResponse `json:"data"`
	Total   int64                    `json:"total"`
	Limit   int                      `json:"limit"`
	Offset  int                      `json:"offset"`
	HasNext bool                     `json:"has_next"`
}
