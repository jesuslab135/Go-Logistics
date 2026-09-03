package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type CreatePurchaseOrderRequest struct {
	// The money columns this record derives are response-only: the server
	// computes them from the line items and the rate terms above, in the same
	// transaction as the write. See internal/domain/money.
	Number             string          `json:"number" binding:"omitempty,max=50"`
	Description        string          `json:"description"`
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
	// state, the workflow timestamps, the actor ids and rejection_reason are
	// response-only. They are moved by POST /purchase-orders/{id}/{action},
	// which is the only thing that can check a transition is legal and record
	// who made it; created_by_id is stamped from the caller.
	Labels       json.RawMessage `json:"labels"`
	CustomFields json.RawMessage `json:"custom_fields" swaggertype:"object"`
}

type UpdatePurchaseOrderRequest struct {
	// The money columns this record derives are response-only: the server
	// computes them from the line items and the rate terms above, in the same
	// transaction as the write. See internal/domain/money.
	Number             string          `json:"number" binding:"omitempty,max=50"`
	Description        string          `json:"description"`
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
	// state, the workflow timestamps, the actor ids and rejection_reason are
	// response-only. They are moved by POST /purchase-orders/{id}/{action},
	// which is the only thing that can check a transition is legal and record
	// who made it; created_by_id is stamped from the caller.
	Labels       json.RawMessage `json:"labels"`
	CustomFields json.RawMessage `json:"custom_fields" swaggertype:"object"`
}

type PurchaseOrderResponse struct {
	ID                  int64            `json:"id"`
	CompanyID           int64            `json:"company_id"`
	Number              string           `json:"number" binding:"omitempty,max=50"`
	Description         string           `json:"description"`
	State               string           `json:"state" binding:"omitempty,max=20"`
	VendorID            int64            `json:"vendor_id"`
	DestinationID       int64            `json:"destination_id"`
	DiscountType        string           `json:"discount_type" binding:"omitempty,max=10"`
	Discount            decimal.Decimal  `json:"discount"`
	DiscountPercentage  decimal.Decimal  `json:"discount_percentage"`
	Tax1Type            string           `json:"tax_1_type" binding:"omitempty,max=10"`
	Tax1                decimal.Decimal  `json:"tax_1"`
	Tax1Percentage      decimal.Decimal  `json:"tax_1_percentage"`
	Tax2Type            string           `json:"tax_2_type" binding:"omitempty,max=10"`
	Tax2                decimal.Decimal  `json:"tax_2"`
	Tax2Percentage      decimal.Decimal  `json:"tax_2_percentage"`
	Shipping            decimal.Decimal  `json:"shipping"`
	Subtotal            decimal.Decimal  `json:"subtotal"`
	DiscountAmount      decimal.Decimal  `json:"discount_amount"`
	Net                 decimal.Decimal  `json:"net"`
	Tax1Amount          decimal.Decimal  `json:"tax_1_amount"`
	Tax2Amount          decimal.Decimal  `json:"tax_2_amount"`
	TotalAmount         decimal.Decimal  `json:"total_amount"`
	TotalOverride       *decimal.Decimal `json:"total_override"`
	TotalOverrideReason *string          `json:"total_override_reason"`
	TotalOverrideByID   *int64           `json:"total_override_by_id"`
	TotalOverrideAt     *time.Time       `json:"total_override_at"`
	CreatedByID         *int64           `json:"created_by_id"`
	SubmittedAt         *time.Time       `json:"submitted_at"`
	SubmittedByID       *int64           `json:"submitted_by_id"`
	RejectedAt          *time.Time       `json:"rejected_at"`
	RejectedByID        *int64           `json:"rejected_by_id"`
	ApprovedAt          *time.Time       `json:"approved_at"`
	ApprovedByID        *int64           `json:"approved_by_id"`
	PurchasedAt         *time.Time       `json:"purchased_at"`
	ReceivedPartialAt   *time.Time       `json:"received_partial_at"`
	ReceivedFullAt      *time.Time       `json:"received_full_at"`
	ClosedAt            *time.Time       `json:"closed_at"`
	Labels              json.RawMessage  `json:"labels"`
	CustomFields        json.RawMessage  `json:"custom_fields" swaggertype:"object"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

type PurchaseOrderPage struct {
	Data    []PurchaseOrderResponse `json:"data"`
	Total   int64                   `json:"total"`
	Limit   int                     `json:"limit"`
	Offset  int                     `json:"offset"`
	HasNext bool                    `json:"has_next"`
}

// PurchaseOrderTransitionRequest is the optional body of a workflow transition.
// The actor is never in it: it comes from the authenticated caller, which is
// the whole point of having transition routes rather than a whole-record PUT
// that let a client name whoever it liked as the approver.
type PurchaseOrderTransitionRequest struct {
	// Reason is required when rejecting and ignored otherwise. It is recorded on
	// the order and in its status log.
	Reason string `json:"reason" binding:"omitempty,max=2000"`
}

// PurchaseOrderStatusLogResponse is one recorded transition. Append-only:
// written by the transition that caused it, inside the same transaction.
type PurchaseOrderStatusLogResponse struct {
	ID              int64     `json:"id"`
	PurchaseOrderID int64     `json:"purchase_order_id"`
	FromState       string    `json:"from_state"`
	ToState         string    `json:"to_state"`
	ActorEmployeeID *int64    `json:"actor_employee_id"`
	ActorType       string    `json:"actor_type"`
	Reason          string    `json:"reason"`
	ChangedAt       time.Time `json:"changed_at"`
}

type PurchaseOrderStatusLogPage struct {
	Data    []PurchaseOrderStatusLogResponse `json:"data"`
	Total   int64                            `json:"total"`
	Limit   int                              `json:"limit"`
	Offset  int                              `json:"offset"`
	HasNext bool                             `json:"has_next"`
}
