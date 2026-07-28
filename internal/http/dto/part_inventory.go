package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreatePartInventoryRequest struct {
	LocationID                 int64            `json:"location_id"`
	AvailableQuantity          decimal.Decimal  `json:"available_quantity"`
	ExpiryDate                 *time.Time       `json:"expiry_date"`
	Aisle                      string           `json:"aisle" binding:"omitempty,max=50"`
	Row                        string           `json:"row" binding:"omitempty,max=50"`
	Bin                        string           `json:"bin" binding:"omitempty,max=50"`
	ReorderPoint               *int32           `json:"reorder_point"`
	ReorderPointEnabled        bool             `json:"reorder_point_enabled"`
	ReorderQuantity            *int32           `json:"reorder_quantity"`
	ReorderPointLeadTimeDays   *int32           `json:"reorder_point_lead_time_days"`
	Active                     *bool            `json:"active"`
	TrackInventory             *bool            `json:"track_inventory"`
	AverageUnitCost            *decimal.Decimal `json:"average_unit_cost"`
	AvailableQuantityUpdatedAt *time.Time       `json:"available_quantity_updated_at"`
}

type UpdatePartInventoryRequest struct {
	LocationID                 int64            `json:"location_id"`
	AvailableQuantity          decimal.Decimal  `json:"available_quantity"`
	ExpiryDate                 *time.Time       `json:"expiry_date"`
	Aisle                      string           `json:"aisle" binding:"omitempty,max=50"`
	Row                        string           `json:"row" binding:"omitempty,max=50"`
	Bin                        string           `json:"bin" binding:"omitempty,max=50"`
	ReorderPoint               *int32           `json:"reorder_point"`
	ReorderPointEnabled        bool             `json:"reorder_point_enabled"`
	ReorderQuantity            *int32           `json:"reorder_quantity"`
	ReorderPointLeadTimeDays   *int32           `json:"reorder_point_lead_time_days"`
	Active                     bool             `json:"active"`
	TrackInventory             bool             `json:"track_inventory"`
	AverageUnitCost            *decimal.Decimal `json:"average_unit_cost"`
	AvailableQuantityUpdatedAt *time.Time       `json:"available_quantity_updated_at"`
}

type PartInventoryResponse struct {
	ID                         int64            `json:"id"`
	PartID                     int64            `json:"part_id"`
	LocationID                 int64            `json:"location_id"`
	AvailableQuantity          decimal.Decimal  `json:"available_quantity"`
	ExpiryDate                 *time.Time       `json:"expiry_date"`
	Aisle                      string           `json:"aisle" binding:"omitempty,max=50"`
	Row                        string           `json:"row" binding:"omitempty,max=50"`
	Bin                        string           `json:"bin" binding:"omitempty,max=50"`
	ReorderPoint               *int32           `json:"reorder_point"`
	ReorderPointEnabled        bool             `json:"reorder_point_enabled"`
	ReorderQuantity            *int32           `json:"reorder_quantity"`
	ReorderPointLeadTimeDays   *int32           `json:"reorder_point_lead_time_days"`
	Active                     bool             `json:"active"`
	TrackInventory             bool             `json:"track_inventory"`
	AverageUnitCost            *decimal.Decimal `json:"average_unit_cost"`
	AvailableQuantityUpdatedAt *time.Time       `json:"available_quantity_updated_at"`
	CreatedAt                  time.Time        `json:"created_at"`
	UpdatedAt                  time.Time        `json:"updated_at"`
}

type PartInventoryPage struct {
	Data    []PartInventoryResponse `json:"data"`
	Total   int64                   `json:"total"`
	Limit   int                     `json:"limit"`
	Offset  int                     `json:"offset"`
	HasNext bool                    `json:"has_next"`
}
