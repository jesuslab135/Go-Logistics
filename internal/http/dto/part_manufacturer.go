package dto

import (
	"time"
)

type CreatePartManufacturerRequest struct {
	Name    string `json:"name" binding:"omitempty,max=200"`
	Website string `json:"website" binding:"omitempty,max=200"`
}

type UpdatePartManufacturerRequest struct {
	Name    string `json:"name" binding:"omitempty,max=200"`
	Website string `json:"website" binding:"omitempty,max=200"`
}

type PartManufacturerResponse struct {
	ID        int64     `json:"id"`
	CompanyID int64     `json:"company_id"`
	Name      string    `json:"name" binding:"omitempty,max=200"`
	Website   string    `json:"website" binding:"omitempty,max=200"`
	CreatedAt time.Time `json:"created_at"`
}

type PartManufacturerPage struct {
	Data    []PartManufacturerResponse `json:"data"`
	Total   int64                      `json:"total"`
	Limit   int                        `json:"limit"`
	Offset  int                        `json:"offset"`
	HasNext bool                       `json:"has_next"`
}
