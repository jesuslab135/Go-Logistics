package inventory

import "time"

// InventoryAdjustmentReason — port of api/models/inventory_model.py:4

type InventoryAdjustmentReason struct {
	ID        int64
	CompanyID int64
	Name      string
	CreatedAt time.Time
}
