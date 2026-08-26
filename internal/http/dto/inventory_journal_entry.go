package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// CreateInventoryJournalEntryRequest files a stock movement. The server applies
// it: part_inventory.available_quantity moves by adjustment_quantity in the same
// transaction that writes the row.
//
// previous_quantity and current_quantity are deliberately absent. They are the
// quantities either side of this adjustment, which only the server can know —
// a client's copy is whatever it last read, and two concurrent adjustments would
// each record a "previous" that was already stale. They are returned, not sent.
type CreateInventoryJournalEntryRequest struct {
	PartID               int64 `json:"part_id"`
	PartLocationDetailID int64 `json:"part_location_detail_id"`
	// AdjustmentQuantity is signed: negative consumes stock, positive receives it.
	AdjustmentQuantity     decimal.Decimal `json:"adjustment_quantity"`
	UnitCost               decimal.Decimal `json:"unit_cost"`
	ReasonID               *int64          `json:"reason_id"`
	WorkOrderID            *int64          `json:"work_order_id"`
	PurchaseOrderLineID    *int64          `json:"purchase_order_line_id"`
	VendorID               *int64          `json:"vendor_id"`
	AdjustmentType         string          `json:"adjustment_type" binding:"omitempty,max=20"`
	TransferPartLocationID *int64          `json:"transfer_part_location_id"`
	Notes                  string          `json:"notes"`
}

// There is no update request: the ledger is append-only. Correct an entry with
// POST /api/v1/inventory-journal-entries/{id}/reverse.

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
	// ReversalOfID names the entry this one offsets, and is null for an ordinary
	// movement. An entry that has been reversed keeps its own values: the
	// correction is a second movement, not an edit of the first.
	ReversalOfID *int64    `json:"reversal_of_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type InventoryJournalEntryPage struct {
	Data    []InventoryJournalEntryResponse `json:"data"`
	Total   int64                           `json:"total"`
	Limit   int                             `json:"limit"`
	Offset  int                             `json:"offset"`
	HasNext bool                            `json:"has_next"`
}
