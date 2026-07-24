package dto

import (
	"time"
)

type CreateMeasurementUnitRequest struct {
	Name         string `json:"name" binding:"omitempty,max=50"`
	Abbreviation string `json:"abbreviation" binding:"omitempty,max=10"`
}

type UpdateMeasurementUnitRequest struct {
	Name         string `json:"name" binding:"omitempty,max=50"`
	Abbreviation string `json:"abbreviation" binding:"omitempty,max=10"`
}

type MeasurementUnitResponse struct {
	ID           int64     `json:"id"`
	CompanyID    int64     `json:"company_id"`
	Name         string    `json:"name" binding:"omitempty,max=50"`
	Abbreviation string    `json:"abbreviation" binding:"omitempty,max=10"`
	CreatedAt    time.Time `json:"created_at"`
}

type MeasurementUnitPage struct {
	Data    []MeasurementUnitResponse `json:"data"`
	Total   int64                     `json:"total"`
	Limit   int                       `json:"limit"`
	Offset  int                       `json:"offset"`
	HasNext bool                      `json:"has_next"`
}
