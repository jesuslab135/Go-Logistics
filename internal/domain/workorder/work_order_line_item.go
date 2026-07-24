package workorder

import (
	"time"

	"github.com/shopspring/decimal"
)

// WorkOrderLineItem — port of api/models/work_order_model.py:283
// ServiceTask is free text despite the name, not an FK (:307).
// M2M: issues -> api_workorderlineitem_issues.

type WorkOrderLineItem struct {
	ID           int64
	WorkOrderID  int64
	LineItemType string
	Title        string
	Description  string
	Position     int32
	ServiceTask  *string
	PartsCost    decimal.Decimal
	LaborCost    decimal.Decimal
	Subtotal     decimal.Decimal
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
