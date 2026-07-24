package serviceentry

import (
	"time"

	"github.com/shopspring/decimal"
)

// ServiceEntry — port of api/models/service_entry_model.py:4
// WorkOrderID is a OneToOne, so the column carries a unique constraint.

type ServiceEntry struct {
	ID                   int64
	CompanyID            int64
	Reference            string
	Status               string
	AssetID              int64
	VendorID             *int64
	WorkOrderID          *int64
	StartedAt            *time.Time
	CompletedAt          *time.Time
	MeterValue           *decimal.Decimal
	PartsSubtotal        decimal.Decimal
	LaborSubtotal        decimal.Decimal
	Subtotal             decimal.Decimal
	Discount             decimal.Decimal
	DiscountType         string
	Tax1                 decimal.Decimal
	Tax1Type             string
	Tax1Percentage       decimal.Decimal
	Tax2                 decimal.Decimal
	Tax2Type             string
	Tax2Percentage       decimal.Decimal
	TotalAmount          decimal.Decimal
	GeneralNotes         string
	IsRoadsideAssistance bool
	LaborTimeSeconds     *int32
	Labels               []string
	CustomFields         map[string]any
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
