package dto

import (
	"time"
)

type CreateFuelTypeRequest struct {
	Name string `json:"name" binding:"omitempty,max=50"`
}

type UpdateFuelTypeRequest struct {
	Name string `json:"name" binding:"omitempty,max=50"`
}

type FuelTypeResponse struct {
	ID        int64     `json:"id"`
	CompanyID int64     `json:"company_id"`
	Name      string    `json:"name" binding:"omitempty,max=50"`
	CreatedAt time.Time `json:"created_at"`
}

type FuelTypePage struct {
	Data    []FuelTypeResponse `json:"data"`
	Total   int64              `json:"total"`
	Limit   int                `json:"limit"`
	Offset  int                `json:"offset"`
	HasNext bool               `json:"has_next"`
}
