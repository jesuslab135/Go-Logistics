package workorder

import (
	"time"

	"github.com/shopspring/decimal"
)

// WorkOrderSubLineItem — port of api/models/work_order_model.py:336
// total_cost was a Python @property (unit_cost * quantity), not a column.

type WorkOrderSubLineItem struct {
	ID                   int64
	LineItemID           int64
	ItemType             string
	Description          string
	Position             int32
	PartID               *int64
	PartLocationDetailID *int64
	TechnicianID         *int64
	UnitCost             decimal.Decimal
	Quantity             decimal.Decimal
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (s WorkOrderSubLineItem) TotalCost() decimal.Decimal {
	return s.UnitCost.Mul(s.Quantity)
}
