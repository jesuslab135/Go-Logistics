package serviceentry

import (
	"time"

	"github.com/shopspring/decimal"
)

// ServiceEntryLineItem — port of api/models/service_entry_model.py:98
// M2M: issues -> api_serviceentrylineitem_issues.

type ServiceEntryLineItem struct {
	ID                int64
	ServiceEntryID    int64
	LineItemType      string
	Description       string
	ServiceTaskID     *int64
	PartID            *int64
	TechnicianID      *int64
	TireID            *int64
	ServiceReminderID *int64
	UnitCost          decimal.Decimal
	Quantity          decimal.Decimal
	PartsCost         decimal.Decimal
	LaborCost         decimal.Decimal
	Subtotal          decimal.Decimal
	Position          int32
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
