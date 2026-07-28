package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type CreatePartRequest struct {
	PartNumber             string           `json:"part_number" binding:"omitempty,max=100"`
	Description            string           `json:"description"`
	UsefulLifeMonths       *int32           `json:"useful_life_months"`
	UsefulLifeDistance     *int32           `json:"useful_life_distance"`
	PartCategoryID         *int64           `json:"part_category_id"`
	PartManufacturerID     *int64           `json:"part_manufacturer_id"`
	MeasurementUnitID      *int64           `json:"measurement_unit_id"`
	ManufacturerPartNumber string           `json:"manufacturer_part_number" binding:"omitempty,max=100"`
	SupplierPartNumber     string           `json:"supplier_part_number" binding:"omitempty,max=100"`
	Upc                    string           `json:"upc" binding:"omitempty,max=50"`
	UnitCost               *decimal.Decimal `json:"unit_cost"`
	InventoryItem          *bool            `json:"inventory_item"`
	ArchivedAt             *time.Time       `json:"archived_at"`
	CustomFields           json.RawMessage  `json:"custom_fields"`
}

type UpdatePartRequest struct {
	PartNumber             string           `json:"part_number" binding:"omitempty,max=100"`
	Description            string           `json:"description"`
	UsefulLifeMonths       *int32           `json:"useful_life_months"`
	UsefulLifeDistance     *int32           `json:"useful_life_distance"`
	PartCategoryID         *int64           `json:"part_category_id"`
	PartManufacturerID     *int64           `json:"part_manufacturer_id"`
	MeasurementUnitID      *int64           `json:"measurement_unit_id"`
	ManufacturerPartNumber string           `json:"manufacturer_part_number" binding:"omitempty,max=100"`
	SupplierPartNumber     string           `json:"supplier_part_number" binding:"omitempty,max=100"`
	Upc                    string           `json:"upc" binding:"omitempty,max=50"`
	UnitCost               *decimal.Decimal `json:"unit_cost"`
	InventoryItem          bool             `json:"inventory_item"`
	ArchivedAt             *time.Time       `json:"archived_at"`
	CustomFields           json.RawMessage  `json:"custom_fields"`
}

type PartResponse struct {
	ID                     int64            `json:"id"`
	CompanyID              int64            `json:"company_id"`
	PartNumber             string           `json:"part_number" binding:"omitempty,max=100"`
	Description            string           `json:"description"`
	UsefulLifeMonths       *int32           `json:"useful_life_months"`
	UsefulLifeDistance     *int32           `json:"useful_life_distance"`
	PartCategoryID         *int64           `json:"part_category_id"`
	PartManufacturerID     *int64           `json:"part_manufacturer_id"`
	MeasurementUnitID      *int64           `json:"measurement_unit_id"`
	ManufacturerPartNumber string           `json:"manufacturer_part_number" binding:"omitempty,max=100"`
	SupplierPartNumber     string           `json:"supplier_part_number" binding:"omitempty,max=100"`
	Upc                    string           `json:"upc" binding:"omitempty,max=50"`
	UnitCost               *decimal.Decimal `json:"unit_cost"`
	InventoryItem          bool             `json:"inventory_item"`
	ArchivedAt             *time.Time       `json:"archived_at"`
	CustomFields           json.RawMessage  `json:"custom_fields"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
}

type PartPage struct {
	Data    []PartResponse `json:"data"`
	Total   int64          `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
	HasNext bool           `json:"has_next"`
}
