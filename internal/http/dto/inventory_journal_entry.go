package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateInventoryJournalEntryRequest struct {
	PartID                 int64           `json:"part_id"`
	PartLocationDetailID   int64           `json:"part_location_detail_id"`
	UserID                 *int64          `json:"user_id"`
	PreviousQuantity       decimal.Decimal `json:"previous_quantity"`
	AdjustmentQuantity     decimal.Decimal `json:"adjustment_quantity"`
	CurrentQuantity        decimal.Decimal `json:"current_quantity"`
	UnitCost               decimal.Decimal `json:"unit_cost"`
	ReasonID               *int64          `json:"reason_id"`
	WorkOrderID            *int64          `json:"work_order_id"`
	PurchaseOrderLineID    *int64          `json:"purchase_order_line_id"`
	VendorID               *int64          `json:"vendor_id"`
	AdjustmentType         string          `json:"adjustment_type" binding:"omitempty,max=20"`
	TransferPartLocationID *int64          `json:"transfer_part_location_id"`
	Notes                  string          `json:"notes"`
}

type UpdateInventoryJournalEntryRequest struct {
	PartID                 int64           `json:"part_id"`
	PartLocationDetailID   int64           `json:"part_location_detail_id"`
	UserID                 *int64          `json:"user_id"`
	PreviousQuantity       decimal.Decimal `json:"previous_quantity"`
	AdjustmentQuantity     decimal.Decimal `json:"adjustment_quantity"`
	CurrentQuantity        decimal.Decimal `json:"current_quantity"`
	UnitCost               decimal.Decimal `json:"unit_cost"`
	ReasonID               *int64          `json:"reason_id"`
	WorkOrderID            *int64          `json:"work_order_id"`
	PurchaseOrderLineID    *int64          `json:"purchase_order_line_id"`
	VendorID               *int64          `json:"vendor_id"`
	AdjustmentType         string          `json:"adjustment_type" binding:"omitempty,max=20"`
	TransferPartLocationID *int64          `json:"transfer_part_location_id"`
	Notes                  string          `json:"notes"`
}

type InventoryJournalEntryResponse struct {
	ID                     int64           `json:"id"`
	CompanyID              int64           `json:"company_id"`
	PartID                 int64           `json:"part_id"`
	PartLocationDetailID   int64           `json:"part_location_detail_id"`
	UserID                 *int64          `json:"user_id"`
	PreviousQuantity       decimal.Decimal `json:"previous_quantity"`
	AdjustmentQuantity     decimal.Decimal `json:"adjustment_quantity"`
	CurrentQuantity        decimal.Decimal `json:"current_quantity"`
	UnitCost               decimal.Decimal `json:"unit_cost"`
	ReasonID               *int64          `json:"reason_id"`
	WorkOrderID            *int64          `json:"work_order_id"`
	PurchaseOrderLineID    *int64          `json:"purchase_order_line_id"`
	VendorID               *int64          `json:"vendor_id"`
	AdjustmentType         string          `json:"adjustment_type" binding:"omitempty,max=20"`
	TransferPartLocationID *int64          `json:"transfer_part_location_id"`
	Notes                  string          `json:"notes"`
	CreatedAt              time.Time       `json:"created_at"`
}

type InventoryJournalEntryPage struct {
	Data    []InventoryJournalEntryResponse `json:"data"`
	Total   int64                           `json:"total"`
	Limit   int                             `json:"limit"`
	Offset  int                             `json:"offset"`
	HasNext bool                            `json:"has_next"`
}
