package dto

import (
	"time"
)

type CreateInventoryAdjustmentReasonRequest struct {
	Name string `json:"name" binding:"omitempty,max=100"`
}

type UpdateInventoryAdjustmentReasonRequest struct {
	Name string `json:"name" binding:"omitempty,max=100"`
}

type InventoryAdjustmentReasonResponse struct {
	ID        int64     `json:"id"`
	CompanyID int64     `json:"company_id"`
	Name      string    `json:"name" binding:"omitempty,max=100"`
	CreatedAt time.Time `json:"created_at"`
}

type InventoryAdjustmentReasonPage struct {
	Data    []InventoryAdjustmentReasonResponse `json:"data"`
	Total   int64                               `json:"total"`
	Limit   int                                 `json:"limit"`
	Offset  int                                 `json:"offset"`
	HasNext bool                                `json:"has_next"`
}
