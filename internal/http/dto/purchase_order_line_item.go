package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreatePurchaseOrderLineItemRequest struct {
	// The money columns this record derives are response-only: the server
	// computes them from the line items and the rate terms above, in the same
	// transaction as the write. See internal/domain/money.
	PartID        int64           `json:"part_id"`
	Quantity      decimal.Decimal `json:"quantity"`
	TotalReceived decimal.Decimal `json:"total_received"`
	UnitCost      decimal.Decimal `json:"unit_cost"`
	Position      int32           `json:"position"`
}

type UpdatePurchaseOrderLineItemRequest struct {
	// The money columns this record derives are response-only: the server
	// computes them from the line items and the rate terms above, in the same
	// transaction as the write. See internal/domain/money.
	PartID        int64           `json:"part_id"`
	Quantity      decimal.Decimal `json:"quantity"`
	TotalReceived decimal.Decimal `json:"total_received"`
	UnitCost      decimal.Decimal `json:"unit_cost"`
	Position      int32           `json:"position"`
}

type PurchaseOrderLineItemResponse struct {
	ID              int64           `json:"id"`
	PurchaseOrderID int64           `json:"purchase_order_id"`
	PartID          int64           `json:"part_id"`
	Quantity        decimal.Decimal `json:"quantity"`
	TotalReceived   decimal.Decimal `json:"total_received"`
	UnitCost        decimal.Decimal `json:"unit_cost"`
	Subtotal        decimal.Decimal `json:"subtotal"`
	Position        int32           `json:"position"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type PurchaseOrderLineItemPage struct {
	Data    []PurchaseOrderLineItemResponse `json:"data"`
	Total   int64                           `json:"total"`
	Limit   int                             `json:"limit"`
	Offset  int                             `json:"offset"`
	HasNext bool                            `json:"has_next"`
}
