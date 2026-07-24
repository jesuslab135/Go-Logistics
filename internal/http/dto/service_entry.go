package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type CreateServiceEntryRequest struct {
	Reference            string           `json:"reference" binding:"omitempty,max=100"`
	Status               string           `json:"status" binding:"omitempty,max=20"`
	AssetID              int64            `json:"asset_id"`
	VendorID             *int64           `json:"vendor_id"`
	WorkOrderID          *int64           `json:"work_order_id"`
	StartedAt            *time.Time       `json:"started_at"`
	CompletedAt          *time.Time       `json:"completed_at"`
	MeterValue           *decimal.Decimal `json:"meter_value"`
	PartsSubtotal        decimal.Decimal  `json:"parts_subtotal"`
	LaborSubtotal        decimal.Decimal  `json:"labor_subtotal"`
	Subtotal             decimal.Decimal  `json:"subtotal"`
	Discount             decimal.Decimal  `json:"discount"`
	DiscountType         string           `json:"discount_type" binding:"omitempty,max=10"`
	Tax1                 decimal.Decimal  `json:"tax_1"`
	Tax1Type             string           `json:"tax_1_type" binding:"omitempty,max=10"`
	Tax1Percentage       decimal.Decimal  `json:"tax_1_percentage"`
	Tax2                 decimal.Decimal  `json:"tax_2"`
	Tax2Type             string           `json:"tax_2_type" binding:"omitempty,max=10"`
	Tax2Percentage       decimal.Decimal  `json:"tax_2_percentage"`
	TotalAmount          decimal.Decimal  `json:"total_amount"`
	GeneralNotes         string           `json:"general_notes"`
	IsRoadsideAssistance bool             `json:"is_roadside_assistance"`
	LaborTimeSeconds     *int32           `json:"labor_time_seconds"`
	Labels               json.RawMessage  `json:"labels"`
	CustomFields         json.RawMessage  `json:"custom_fields"`
}

type UpdateServiceEntryRequest struct {
	Reference            string           `json:"reference" binding:"omitempty,max=100"`
	Status               string           `json:"status" binding:"omitempty,max=20"`
	AssetID              int64            `json:"asset_id"`
	VendorID             *int64           `json:"vendor_id"`
	WorkOrderID          *int64           `json:"work_order_id"`
	StartedAt            *time.Time       `json:"started_at"`
	CompletedAt          *time.Time       `json:"completed_at"`
	MeterValue           *decimal.Decimal `json:"meter_value"`
	PartsSubtotal        decimal.Decimal  `json:"parts_subtotal"`
	LaborSubtotal        decimal.Decimal  `json:"labor_subtotal"`
	Subtotal             decimal.Decimal  `json:"subtotal"`
	Discount             decimal.Decimal  `json:"discount"`
	DiscountType         string           `json:"discount_type" binding:"omitempty,max=10"`
	Tax1                 decimal.Decimal  `json:"tax_1"`
	Tax1Type             string           `json:"tax_1_type" binding:"omitempty,max=10"`
	Tax1Percentage       decimal.Decimal  `json:"tax_1_percentage"`
	Tax2                 decimal.Decimal  `json:"tax_2"`
	Tax2Type             string           `json:"tax_2_type" binding:"omitempty,max=10"`
	Tax2Percentage       decimal.Decimal  `json:"tax_2_percentage"`
	TotalAmount          decimal.Decimal  `json:"total_amount"`
	GeneralNotes         string           `json:"general_notes"`
	IsRoadsideAssistance bool             `json:"is_roadside_assistance"`
	LaborTimeSeconds     *int32           `json:"labor_time_seconds"`
	Labels               json.RawMessage  `json:"labels"`
	CustomFields         json.RawMessage  `json:"custom_fields"`
}

type ServiceEntryResponse struct {
	ID                   int64            `json:"id"`
	CompanyID            int64            `json:"company_id"`
	Reference            string           `json:"reference" binding:"omitempty,max=100"`
	Status               string           `json:"status" binding:"omitempty,max=20"`
	AssetID              int64            `json:"asset_id"`
	VendorID             *int64           `json:"vendor_id"`
	WorkOrderID          *int64           `json:"work_order_id"`
	StartedAt            *time.Time       `json:"started_at"`
	CompletedAt          *time.Time       `json:"completed_at"`
	MeterValue           *decimal.Decimal `json:"meter_value"`
	PartsSubtotal        decimal.Decimal  `json:"parts_subtotal"`
	LaborSubtotal        decimal.Decimal  `json:"labor_subtotal"`
	Subtotal             decimal.Decimal  `json:"subtotal"`
	Discount             decimal.Decimal  `json:"discount"`
	DiscountType         string           `json:"discount_type" binding:"omitempty,max=10"`
	Tax1                 decimal.Decimal  `json:"tax_1"`
	Tax1Type             string           `json:"tax_1_type" binding:"omitempty,max=10"`
	Tax1Percentage       decimal.Decimal  `json:"tax_1_percentage"`
	Tax2                 decimal.Decimal  `json:"tax_2"`
	Tax2Type             string           `json:"tax_2_type" binding:"omitempty,max=10"`
	Tax2Percentage       decimal.Decimal  `json:"tax_2_percentage"`
	TotalAmount          decimal.Decimal  `json:"total_amount"`
	GeneralNotes         string           `json:"general_notes"`
	IsRoadsideAssistance bool             `json:"is_roadside_assistance"`
	LaborTimeSeconds     *int32           `json:"labor_time_seconds"`
	Labels               json.RawMessage  `json:"labels"`
	CustomFields         json.RawMessage  `json:"custom_fields"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

type ServiceEntryPage struct {
	Data    []ServiceEntryResponse `json:"data"`
	Total   int64                  `json:"total"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
	HasNext bool                   `json:"has_next"`
}
