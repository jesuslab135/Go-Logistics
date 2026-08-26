package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateWorkOrderLineItemRequest struct {
	// The money columns this record derives are response-only: the server
	// computes them from the line items and the rate terms above, in the same
	// transaction as the write. See internal/domain/money.
	LineItemType string  `json:"line_item_type" binding:"omitempty,max=20"`
	Title        string  `json:"title" binding:"omitempty,max=255"`
	Description  string  `json:"description"`
	Position     int32   `json:"position"`
	ServiceTask  *string `json:"service_task" binding:"omitempty,max=255"`
}

type UpdateWorkOrderLineItemRequest struct {
	// The money columns this record derives are response-only: the server
	// computes them from the line items and the rate terms above, in the same
	// transaction as the write. See internal/domain/money.
	LineItemType string  `json:"line_item_type" binding:"omitempty,max=20"`
	Title        string  `json:"title" binding:"omitempty,max=255"`
	Description  string  `json:"description"`
	Position     int32   `json:"position"`
	ServiceTask  *string `json:"service_task" binding:"omitempty,max=255"`
}

type WorkOrderLineItemResponse struct {
	ID           int64           `json:"id"`
	WorkOrderID  int64           `json:"work_order_id"`
	LineItemType string          `json:"line_item_type" binding:"omitempty,max=20"`
	Title        string          `json:"title" binding:"omitempty,max=255"`
	Description  string          `json:"description"`
	Position     int32           `json:"position"`
	ServiceTask  *string         `json:"service_task" binding:"omitempty,max=255"`
	PartsCost    decimal.Decimal `json:"parts_cost"`
	LaborCost    decimal.Decimal `json:"labor_cost"`
	Subtotal     decimal.Decimal `json:"subtotal"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type WorkOrderLineItemPage struct {
	Data    []WorkOrderLineItemResponse `json:"data"`
	Total   int64                       `json:"total"`
	Limit   int                         `json:"limit"`
	Offset  int                         `json:"offset"`
	HasNext bool                        `json:"has_next"`
}
