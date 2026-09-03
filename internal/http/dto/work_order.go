package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type CreateWorkOrderRequest struct {
	// The money columns this record derives are response-only: the server
	// computes them from the line items and the rate terms above, in the same
	// transaction as the write. See internal/domain/money.
	LocationID            *int64           `json:"location_id"`
	Number                string           `json:"number" binding:"omitempty,max=50"`
	Description           string           `json:"description"`
	AssetID               int64            `json:"asset_id"`
	StatusID              int64            `json:"status_id"`
	VendorID              *int64           `json:"vendor_id"`
	AssignedToID          *int64           `json:"assigned_to_id"`
	IssuedByID            *int64           `json:"issued_by_id"`
	FaultID               *int64           `json:"fault_id"`
	IssuedAt              time.Time        `json:"issued_at"`
	ScheduledAt           *time.Time       `json:"scheduled_at"`
	StartedAt             *time.Time       `json:"started_at"`
	ExpectedCompletedAt   *time.Time       `json:"expected_completed_at"`
	CompletedAt           *time.Time       `json:"completed_at"`
	StartingMeter         *decimal.Decimal `json:"starting_meter"`
	EndingMeter           *decimal.Decimal `json:"ending_meter"`
	DurationSeconds       *int32           `json:"duration_seconds"`
	LaborTimeSeconds      *int32           `json:"labor_time_seconds"`
	PartsMarkupType       string           `json:"parts_markup_type" binding:"omitempty,max=10"`
	PartsMarkup           decimal.Decimal  `json:"parts_markup"`
	PartsMarkupPercentage decimal.Decimal  `json:"parts_markup_percentage"`
	LaborMarkupType       string           `json:"labor_markup_type" binding:"omitempty,max=10"`
	LaborMarkup           decimal.Decimal  `json:"labor_markup"`
	LaborMarkupPercentage decimal.Decimal  `json:"labor_markup_percentage"`
	Discount              decimal.Decimal  `json:"discount"`
	DiscountType          string           `json:"discount_type" binding:"omitempty,max=10"`
	Tax1                  decimal.Decimal  `json:"tax_1"`
	Tax1Type              string           `json:"tax_1_type" binding:"omitempty,max=10"`
	Tax1Percentage        decimal.Decimal  `json:"tax_1_percentage"`
	Tax2                  decimal.Decimal  `json:"tax_2"`
	Tax2Type              string           `json:"tax_2_type" binding:"omitempty,max=10"`
	Tax2Percentage        decimal.Decimal  `json:"tax_2_percentage"`
	InvoiceNumber         string           `json:"invoice_number" binding:"omitempty,max=100"`
	PurchaseOrderNumber   string           `json:"purchase_order_number" binding:"omitempty,max=100"`
	CommentsCount         int32            `json:"comments_count"`
	ImagesCount           int32            `json:"images_count"`
	DocumentsCount        int32            `json:"documents_count"`
	Labels                json.RawMessage  `json:"labels"`
	CustomFields          json.RawMessage  `json:"custom_fields" swaggertype:"object"`
}

type UpdateWorkOrderRequest struct {
	// The money columns this record derives are response-only: the server
	// computes them from the line items and the rate terms above, in the same
	// transaction as the write. See internal/domain/money.
	LocationID            *int64           `json:"location_id"`
	Number                string           `json:"number" binding:"omitempty,max=50"`
	Description           string           `json:"description"`
	AssetID               int64            `json:"asset_id"`
	StatusID              int64            `json:"status_id"`
	VendorID              *int64           `json:"vendor_id"`
	AssignedToID          *int64           `json:"assigned_to_id"`
	IssuedByID            *int64           `json:"issued_by_id"`
	FaultID               *int64           `json:"fault_id"`
	IssuedAt              time.Time        `json:"issued_at"`
	ScheduledAt           *time.Time       `json:"scheduled_at"`
	StartedAt             *time.Time       `json:"started_at"`
	ExpectedCompletedAt   *time.Time       `json:"expected_completed_at"`
	CompletedAt           *time.Time       `json:"completed_at"`
	StartingMeter         *decimal.Decimal `json:"starting_meter"`
	EndingMeter           *decimal.Decimal `json:"ending_meter"`
	DurationSeconds       *int32           `json:"duration_seconds"`
	LaborTimeSeconds      *int32           `json:"labor_time_seconds"`
	PartsMarkupType       string           `json:"parts_markup_type" binding:"omitempty,max=10"`
	PartsMarkup           decimal.Decimal  `json:"parts_markup"`
	PartsMarkupPercentage decimal.Decimal  `json:"parts_markup_percentage"`
	LaborMarkupType       string           `json:"labor_markup_type" binding:"omitempty,max=10"`
	LaborMarkup           decimal.Decimal  `json:"labor_markup"`
	LaborMarkupPercentage decimal.Decimal  `json:"labor_markup_percentage"`
	Discount              decimal.Decimal  `json:"discount"`
	DiscountType          string           `json:"discount_type" binding:"omitempty,max=10"`
	Tax1                  decimal.Decimal  `json:"tax_1"`
	Tax1Type              string           `json:"tax_1_type" binding:"omitempty,max=10"`
	Tax1Percentage        decimal.Decimal  `json:"tax_1_percentage"`
	Tax2                  decimal.Decimal  `json:"tax_2"`
	Tax2Type              string           `json:"tax_2_type" binding:"omitempty,max=10"`
	Tax2Percentage        decimal.Decimal  `json:"tax_2_percentage"`
	InvoiceNumber         string           `json:"invoice_number" binding:"omitempty,max=100"`
	PurchaseOrderNumber   string           `json:"purchase_order_number" binding:"omitempty,max=100"`
	CommentsCount         int32            `json:"comments_count"`
	ImagesCount           int32            `json:"images_count"`
	DocumentsCount        int32            `json:"documents_count"`
	Labels                json.RawMessage  `json:"labels"`
	CustomFields          json.RawMessage  `json:"custom_fields" swaggertype:"object"`
}

type WorkOrderResponse struct {
	ID                    int64            `json:"id"`
	LocationID            *int64           `json:"location_id"`
	CompanyID             int64            `json:"company_id"`
	Number                string           `json:"number" binding:"omitempty,max=50"`
	Description           string           `json:"description"`
	AssetID               int64            `json:"asset_id"`
	StatusID              int64            `json:"status_id"`
	VendorID              *int64           `json:"vendor_id"`
	AssignedToID          *int64           `json:"assigned_to_id"`
	IssuedByID            *int64           `json:"issued_by_id"`
	FaultID               *int64           `json:"fault_id"`
	IssuedAt              time.Time        `json:"issued_at"`
	ScheduledAt           *time.Time       `json:"scheduled_at"`
	StartedAt             *time.Time       `json:"started_at"`
	ExpectedCompletedAt   *time.Time       `json:"expected_completed_at"`
	CompletedAt           *time.Time       `json:"completed_at"`
	StartingMeter         *decimal.Decimal `json:"starting_meter"`
	EndingMeter           *decimal.Decimal `json:"ending_meter"`
	DurationSeconds       *int32           `json:"duration_seconds"`
	LaborTimeSeconds      *int32           `json:"labor_time_seconds"`
	PartsMarkupType       string           `json:"parts_markup_type" binding:"omitempty,max=10"`
	PartsMarkup           decimal.Decimal  `json:"parts_markup"`
	PartsMarkupPercentage decimal.Decimal  `json:"parts_markup_percentage"`
	LaborMarkupType       string           `json:"labor_markup_type" binding:"omitempty,max=10"`
	LaborMarkup           decimal.Decimal  `json:"labor_markup"`
	LaborMarkupPercentage decimal.Decimal  `json:"labor_markup_percentage"`
	PartsSubtotal         decimal.Decimal  `json:"parts_subtotal"`
	LaborSubtotal         decimal.Decimal  `json:"labor_subtotal"`
	Subtotal              decimal.Decimal  `json:"subtotal"`
	Discount              decimal.Decimal  `json:"discount"`
	DiscountType          string           `json:"discount_type" binding:"omitempty,max=10"`
	DiscountPercentage    decimal.Decimal  `json:"discount_percentage"`
	DiscountAmount        decimal.Decimal  `json:"discount_amount"`
	Net                   decimal.Decimal  `json:"net"`
	Tax1                  decimal.Decimal  `json:"tax_1"`
	Tax1Type              string           `json:"tax_1_type" binding:"omitempty,max=10"`
	Tax1Percentage        decimal.Decimal  `json:"tax_1_percentage"`
	Tax1Amount            decimal.Decimal  `json:"tax_1_amount"`
	Tax2                  decimal.Decimal  `json:"tax_2"`
	Tax2Type              string           `json:"tax_2_type" binding:"omitempty,max=10"`
	Tax2Percentage        decimal.Decimal  `json:"tax_2_percentage"`
	Tax2Amount            decimal.Decimal  `json:"tax_2_amount"`
	TotalAmount           decimal.Decimal  `json:"total_amount"`
	TotalOverride         *decimal.Decimal `json:"total_override"`
	TotalOverrideReason   *string          `json:"total_override_reason"`
	TotalOverrideByID     *int64           `json:"total_override_by_id"`
	TotalOverrideAt       *time.Time       `json:"total_override_at"`
	InvoiceNumber         string           `json:"invoice_number" binding:"omitempty,max=100"`
	PurchaseOrderNumber   string           `json:"purchase_order_number" binding:"omitempty,max=100"`
	CommentsCount         int32            `json:"comments_count"`
	ImagesCount           int32            `json:"images_count"`
	DocumentsCount        int32            `json:"documents_count"`
	Labels                json.RawMessage  `json:"labels"`
	CustomFields          json.RawMessage  `json:"custom_fields" swaggertype:"object"`
	CreatedAt             time.Time        `json:"created_at"`
	UpdatedAt             time.Time        `json:"updated_at"`
}

type WorkOrderPage struct {
	Data    []WorkOrderResponse `json:"data"`
	Total   int64               `json:"total"`
	Limit   int                 `json:"limit"`
	Offset  int                 `json:"offset"`
	HasNext bool                `json:"has_next"`
}
