package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateWorkOrderLineItemRequest struct {
	LineItemType string          `json:"line_item_type" binding:"omitempty,max=20"`
	Title        string          `json:"title" binding:"omitempty,max=255"`
	Description  string          `json:"description"`
	Position     int32           `json:"position"`
	ServiceTask  *string         `json:"service_task" binding:"omitempty,max=255"`
	PartsCost    decimal.Decimal `json:"parts_cost"`
	LaborCost    decimal.Decimal `json:"labor_cost"`
	Subtotal     decimal.Decimal `json:"subtotal"`
}

type UpdateWorkOrderLineItemRequest struct {
	LineItemType string          `json:"line_item_type" binding:"omitempty,max=20"`
	Title        string          `json:"title" binding:"omitempty,max=255"`
	Description  string          `json:"description"`
	Position     int32           `json:"position"`
	ServiceTask  *string         `json:"service_task" binding:"omitempty,max=255"`
	PartsCost    decimal.Decimal `json:"parts_cost"`
	LaborCost    decimal.Decimal `json:"labor_cost"`
	Subtotal     decimal.Decimal `json:"subtotal"`
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
