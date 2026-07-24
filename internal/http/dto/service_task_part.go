package dto

import (
	"github.com/shopspring/decimal"
)

type CreateServiceTaskPartRequest struct {
	PartID   int64           `json:"part_id"`
	Quantity decimal.Decimal `json:"quantity"`
	Position int32           `json:"position"`
}

type UpdateServiceTaskPartRequest struct {
	PartID   int64           `json:"part_id"`
	Quantity decimal.Decimal `json:"quantity"`
	Position int32           `json:"position"`
}

type ServiceTaskPartResponse struct {
	ID            int64           `json:"id"`
	ServiceTaskID int64           `json:"service_task_id"`
	PartID        int64           `json:"part_id"`
	Quantity      decimal.Decimal `json:"quantity"`
	Position      int32           `json:"position"`
}

type ServiceTaskPartPage struct {
	Data    []ServiceTaskPartResponse `json:"data"`
	Total   int64                     `json:"total"`
	Limit   int                       `json:"limit"`
	Offset  int                       `json:"offset"`
	HasNext bool                      `json:"has_next"`
}
