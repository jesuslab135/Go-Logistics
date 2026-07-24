package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type CreatePurchaseOrderRequest struct {
	Number             string          `json:"number" binding:"omitempty,max=50"`
	Description        string          `json:"description"`
	State              string          `json:"state" binding:"omitempty,max=20"`
	VendorID           int64           `json:"vendor_id"`
	DestinationID      int64           `json:"destination_id"`
	DiscountType       string          `json:"discount_type" binding:"omitempty,max=10"`
	Discount           decimal.Decimal `json:"discount"`
	DiscountPercentage decimal.Decimal `json:"discount_percentage"`
	Tax1Type           string          `json:"tax_1_type" binding:"omitempty,max=10"`
	Tax1               decimal.Decimal `json:"tax_1"`
	Tax1Percentage     decimal.Decimal `json:"tax_1_percentage"`
	Tax2Type           string          `json:"tax_2_type" binding:"omitempty,max=10"`
	Tax2               decimal.Decimal `json:"tax_2"`
	Tax2Percentage     decimal.Decimal `json:"tax_2_percentage"`
	Shipping           decimal.Decimal `json:"shipping"`
	Subtotal           decimal.Decimal `json:"subtotal"`
	TotalAmount        decimal.Decimal `json:"total_amount"`
	CreatedByID        *int64          `json:"created_by_id"`
	SubmittedAt        *time.Time      `json:"submitted_at"`
	SubmittedByID      *int64          `json:"submitted_by_id"`
	RejectedAt         *time.Time      `json:"rejected_at"`
	RejectedByID       *int64          `json:"rejected_by_id"`
	ApprovedAt         *time.Time      `json:"approved_at"`
	ApprovedByID       *int64          `json:"approved_by_id"`
	PurchasedAt        *time.Time      `json:"purchased_at"`
	ReceivedPartialAt  *time.Time      `json:"received_partial_at"`
	ReceivedFullAt     *time.Time      `json:"received_full_at"`
	ClosedAt           *time.Time      `json:"closed_at"`
	Labels             json.RawMessage `json:"labels"`
	CustomFields       json.RawMessage `json:"custom_fields"`
}

type UpdatePurchaseOrderRequest struct {
	Number             string          `json:"number" binding:"omitempty,max=50"`
	Description        string          `json:"description"`
	State              string          `json:"state" binding:"omitempty,max=20"`
	VendorID           int64           `json:"vendor_id"`
	DestinationID      int64           `json:"destination_id"`
	DiscountType       string          `json:"discount_type" binding:"omitempty,max=10"`
	Discount           decimal.Decimal `json:"discount"`
	DiscountPercentage decimal.Decimal `json:"discount_percentage"`
	Tax1Type           string          `json:"tax_1_type" binding:"omitempty,max=10"`
	Tax1               decimal.Decimal `json:"tax_1"`
	Tax1Percentage     decimal.Decimal `json:"tax_1_percentage"`
	Tax2Type           string          `json:"tax_2_type" binding:"omitempty,max=10"`
	Tax2               decimal.Decimal `json:"tax_2"`
	Tax2Percentage     decimal.Decimal `json:"tax_2_percentage"`
	Shipping           decimal.Decimal `json:"shipping"`
	Subtotal           decimal.Decimal `json:"subtotal"`
	TotalAmount        decimal.Decimal `json:"total_amount"`
	CreatedByID        *int64          `json:"created_by_id"`
	SubmittedAt        *time.Time      `json:"submitted_at"`
	SubmittedByID      *int64          `json:"submitted_by_id"`
	RejectedAt         *time.Time      `json:"rejected_at"`
	RejectedByID       *int64          `json:"rejected_by_id"`
	ApprovedAt         *time.Time      `json:"approved_at"`
	ApprovedByID       *int64          `json:"approved_by_id"`
	PurchasedAt        *time.Time      `json:"purchased_at"`
	ReceivedPartialAt  *time.Time      `json:"received_partial_at"`
	ReceivedFullAt     *time.Time      `json:"received_full_at"`
	ClosedAt           *time.Time      `json:"closed_at"`
	Labels             json.RawMessage `json:"labels"`
	CustomFields       json.RawMessage `json:"custom_fields"`
}

type PurchaseOrderResponse struct {
	ID                 int64           `json:"id"`
	CompanyID          int64           `json:"company_id"`
	Number             string          `json:"number" binding:"omitempty,max=50"`
	Description        string          `json:"description"`
	State              string          `json:"state" binding:"omitempty,max=20"`
	VendorID           int64           `json:"vendor_id"`
	DestinationID      int64           `json:"destination_id"`
	DiscountType       string          `json:"discount_type" binding:"omitempty,max=10"`
	Discount           decimal.Decimal `json:"discount"`
	DiscountPercentage decimal.Decimal `json:"discount_percentage"`
	Tax1Type           string          `json:"tax_1_type" binding:"omitempty,max=10"`
	Tax1               decimal.Decimal `json:"tax_1"`
	Tax1Percentage     decimal.Decimal `json:"tax_1_percentage"`
	Tax2Type           string          `json:"tax_2_type" binding:"omitempty,max=10"`
	Tax2               decimal.Decimal `json:"tax_2"`
	Tax2Percentage     decimal.Decimal `json:"tax_2_percentage"`
	Shipping           decimal.Decimal `json:"shipping"`
	Subtotal           decimal.Decimal `json:"subtotal"`
	TotalAmount        decimal.Decimal `json:"total_amount"`
	CreatedByID        *int64          `json:"created_by_id"`
	SubmittedAt        *time.Time      `json:"submitted_at"`
	SubmittedByID      *int64          `json:"submitted_by_id"`
	RejectedAt         *time.Time      `json:"rejected_at"`
	RejectedByID       *int64          `json:"rejected_by_id"`
	ApprovedAt         *time.Time      `json:"approved_at"`
	ApprovedByID       *int64          `json:"approved_by_id"`
	PurchasedAt        *time.Time      `json:"purchased_at"`
	ReceivedPartialAt  *time.Time      `json:"received_partial_at"`
	ReceivedFullAt     *time.Time      `json:"received_full_at"`
	ClosedAt           *time.Time      `json:"closed_at"`
	Labels             json.RawMessage `json:"labels"`
	CustomFields       json.RawMessage `json:"custom_fields"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type PurchaseOrderPage struct {
	Data    []PurchaseOrderResponse `json:"data"`
	Total   int64                   `json:"total"`
	Limit   int                     `json:"limit"`
	Offset  int                     `json:"offset"`
	HasNext bool                    `json:"has_next"`
}
