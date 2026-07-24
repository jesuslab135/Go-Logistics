package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateWorkOrderSubLineItemRequest struct {
	ItemType             string          `json:"item_type" binding:"omitempty,max=10"`
	Description          string          `json:"description" binding:"omitempty,max=255"`
	Position             int32           `json:"position"`
	PartID               *int64          `json:"part_id"`
	PartLocationDetailID *int64          `json:"part_location_detail_id"`
	TechnicianID         *int64          `json:"technician_id"`
	UnitCost             decimal.Decimal `json:"unit_cost"`
	Quantity             decimal.Decimal `json:"quantity"`
}

type UpdateWorkOrderSubLineItemRequest struct {
	ItemType             string          `json:"item_type" binding:"omitempty,max=10"`
	Description          string          `json:"description" binding:"omitempty,max=255"`
	Position             int32           `json:"position"`
	PartID               *int64          `json:"part_id"`
	PartLocationDetailID *int64          `json:"part_location_detail_id"`
	TechnicianID         *int64          `json:"technician_id"`
	UnitCost             decimal.Decimal `json:"unit_cost"`
	Quantity             decimal.Decimal `json:"quantity"`
}

type WorkOrderSubLineItemResponse struct {
	ID                   int64           `json:"id"`
	LineItemID           int64           `json:"line_item_id"`
	ItemType             string          `json:"item_type" binding:"omitempty,max=10"`
	Description          string          `json:"description" binding:"omitempty,max=255"`
	Position             int32           `json:"position"`
	PartID               *int64          `json:"part_id"`
	PartLocationDetailID *int64          `json:"part_location_detail_id"`
	TechnicianID         *int64          `json:"technician_id"`
	UnitCost             decimal.Decimal `json:"unit_cost"`
	Quantity             decimal.Decimal `json:"quantity"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

type WorkOrderSubLineItemPage struct {
	Data    []WorkOrderSubLineItemResponse `json:"data"`
	Total   int64                          `json:"total"`
	Limit   int                            `json:"limit"`
	Offset  int                            `json:"offset"`
	HasNext bool                           `json:"has_next"`
}
