package purchaseorder

import (
	"time"

	"github.com/shopspring/decimal"
)

// PurchaseOrder — port of api/models/purchase_order_model.py:4
// DestinationID references PartLocation.

type PurchaseOrder struct {
	ID                 int64
	CompanyID          int64
	Number             string
	Description        string
	State              string
	VendorID           int64
	DestinationID      int64
	DiscountType       string
	Discount           decimal.Decimal
	DiscountPercentage decimal.Decimal
	Tax1Type           string
	Tax1               decimal.Decimal
	Tax1Percentage     decimal.Decimal
	Tax2Type           string
	Tax2               decimal.Decimal
	Tax2Percentage     decimal.Decimal
	Shipping           decimal.Decimal
	Subtotal           decimal.Decimal
	TotalAmount        decimal.Decimal
	CreatedByID        *int64
	SubmittedAt        *time.Time
	SubmittedByID      *int64
	RejectedAt         *time.Time
	RejectedByID       *int64
	ApprovedAt         *time.Time
	ApprovedByID       *int64
	PurchasedAt        *time.Time
	ReceivedPartialAt  *time.Time
	ReceivedFullAt     *time.Time
	ClosedAt           *time.Time
	Labels             []string
	CustomFields       map[string]any
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
