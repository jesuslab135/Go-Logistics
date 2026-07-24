package workorder

import (
	"time"

	"github.com/shopspring/decimal"
)

// WorkOrder — port of api/models/work_order_model.py:60
// Number is generated in Django's save() under SELECT ... FOR UPDATE (:259);
// that sequencing belongs in a service, not here.
// M2M: issues -> api_workorder_issues, faults -> api_workorder_faults.

type WorkOrder struct {
	ID                    int64
	LocationID            *int64
	CompanyID             int64
	Number                string
	Description           string
	AssetID               int64
	StatusID              int64
	VendorID              *int64
	AssignedToID          *int64
	IssuedByID            *int64
	FaultID               *int64
	IssuedAt              time.Time
	ScheduledAt           *time.Time
	StartedAt             *time.Time
	ExpectedCompletedAt   *time.Time
	CompletedAt           *time.Time
	StartingMeter         *decimal.Decimal
	EndingMeter           *decimal.Decimal
	DurationSeconds       *int32
	LaborTimeSeconds      *int32
	PartsMarkupType       string
	PartsMarkup           decimal.Decimal
	PartsMarkupPercentage decimal.Decimal
	LaborMarkupType       string
	LaborMarkup           decimal.Decimal
	LaborMarkupPercentage decimal.Decimal
	PartsSubtotal         decimal.Decimal
	LaborSubtotal         decimal.Decimal
	Subtotal              decimal.Decimal
	Discount              decimal.Decimal
	DiscountType          string
	Tax1                  decimal.Decimal
	Tax1Type              string
	Tax1Percentage        decimal.Decimal
	Tax2                  decimal.Decimal
	Tax2Type              string
	Tax2Percentage        decimal.Decimal
	TotalAmount           decimal.Decimal
	InvoiceNumber         string
	PurchaseOrderNumber   string
	CommentsCount         int32
	ImagesCount           int32
	DocumentsCount        int32
	Labels                []string
	CustomFields          map[string]any
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
