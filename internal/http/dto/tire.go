package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateTireRequest struct {
	TireIdentificationNumber string           `json:"tire_identification_number" binding:"omitempty,max=100"`
	TireModelID              *int64           `json:"tire_model_id"`
	Status                   string           `json:"status" binding:"omitempty,max=15"`
	CurrentTreadDepth32nds   *int32           `json:"current_tread_depth_32nds"`
	CurrentPsi               *decimal.Decimal `json:"current_psi"`
	TotalMiles               int32            `json:"total_miles"`
	CurrentVehicleID         *int64           `json:"current_vehicle_id"`
	CurrentPositionCode      string           `json:"current_position_code" binding:"omitempty,max=10"`
	PurchaseDate             *time.Time       `json:"purchase_date"`
	PurchaseCost             *decimal.Decimal `json:"purchase_cost"`
	VendorID                 *int64           `json:"vendor_id"`
}

type UpdateTireRequest struct {
	TireIdentificationNumber string           `json:"tire_identification_number" binding:"omitempty,max=100"`
	TireModelID              *int64           `json:"tire_model_id"`
	Status                   string           `json:"status" binding:"omitempty,max=15"`
	CurrentTreadDepth32nds   *int32           `json:"current_tread_depth_32nds"`
	CurrentPsi               *decimal.Decimal `json:"current_psi"`
	TotalMiles               int32            `json:"total_miles"`
	CurrentVehicleID         *int64           `json:"current_vehicle_id"`
	CurrentPositionCode      string           `json:"current_position_code" binding:"omitempty,max=10"`
	PurchaseDate             *time.Time       `json:"purchase_date"`
	PurchaseCost             *decimal.Decimal `json:"purchase_cost"`
	VendorID                 *int64           `json:"vendor_id"`
}

type TireResponse struct {
	ID                       int64            `json:"id"`
	CompanyID                int64            `json:"company_id"`
	TireIdentificationNumber string           `json:"tire_identification_number" binding:"omitempty,max=100"`
	TireModelID              *int64           `json:"tire_model_id"`
	Status                   string           `json:"status" binding:"omitempty,max=15"`
	CurrentTreadDepth32nds   *int32           `json:"current_tread_depth_32nds"`
	CurrentPsi               *decimal.Decimal `json:"current_psi"`
	TotalMiles               int32            `json:"total_miles"`
	CurrentVehicleID         *int64           `json:"current_vehicle_id"`
	CurrentPositionCode      string           `json:"current_position_code" binding:"omitempty,max=10"`
	PurchaseDate             *time.Time       `json:"purchase_date"`
	PurchaseCost             *decimal.Decimal `json:"purchase_cost"`
	VendorID                 *int64           `json:"vendor_id"`
	CreatedAt                time.Time        `json:"created_at"`
}

type TirePage struct {
	Data    []TireResponse `json:"data"`
	Total   int64          `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
	HasNext bool           `json:"has_next"`
}
