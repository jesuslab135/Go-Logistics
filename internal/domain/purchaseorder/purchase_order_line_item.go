package purchaseorder

import (
	"time"

	"github.com/shopspring/decimal"
)

// PurchaseOrderLineItem — port of api/models/purchase_order_model.py:96

type PurchaseOrderLineItem struct {
	ID              int64
	PurchaseOrderID int64
	PartID          int64
	Quantity        decimal.Decimal
	TotalReceived   decimal.Decimal
	UnitCost        decimal.Decimal
	Subtotal        decimal.Decimal
	Position        int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
