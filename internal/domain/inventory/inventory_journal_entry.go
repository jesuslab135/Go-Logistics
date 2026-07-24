package inventory

import (
	"time"

	"github.com/shopspring/decimal"
)

// InventoryJournalEntry — port of api/models/inventory_model.py:21
// Immutable audit trail; UserID references Employee, not auth_user.

type InventoryJournalEntry struct {
	ID                     int64
	CompanyID              int64
	PartID                 int64
	PartLocationDetailID   int64
	UserID                 *int64
	PreviousQuantity       decimal.Decimal
	AdjustmentQuantity     decimal.Decimal
	CurrentQuantity        decimal.Decimal
	UnitCost               decimal.Decimal
	ReasonID               *int64
	WorkOrderID            *int64
	PurchaseOrderLineID    *int64
	VendorID               *int64
	AdjustmentType         string
	TransferPartLocationID *int64
	Notes                  string
	CreatedAt              time.Time
}
