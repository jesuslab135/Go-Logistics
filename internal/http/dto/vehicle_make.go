package dto

import "time"

// --- VehicleMake ---

type CreateVehicleMakeRequest struct {
	Name string `json:"name" binding:"required,max=100"`
}

type UpdateVehicleMakeRequest struct {
	Name string `json:"name" binding:"required,max=100"`
}

type VehicleMakeResponse struct {
	ID        int64     `json:"id"`
	CompanyID int64     `json:"company_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VehicleMakePage struct {
	Data    []VehicleMakeResponse `json:"data"`
	Total   int64                 `json:"total"`
	Limit   int                   `json:"limit"`
	Offset  int                   `json:"offset"`
	HasNext bool                  `json:"has_next"`
}

// --- VehicleModel ---

type CreateVehicleModelRequest struct {
	Name   string `json:"name" binding:"required,max=100"`
	MakeID *int64 `json:"make_id"`
}

type UpdateVehicleModelRequest struct {
	Name   string `json:"name" binding:"required,max=100"`
	MakeID *int64 `json:"make_id"`
}

type VehicleModelResponse struct {
	ID        int64     `json:"id"`
	CompanyID int64     `json:"company_id"`
	Name      string    `json:"name"`
	MakeID    *int64    `json:"make_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VehicleModelPage struct {
	Data    []VehicleModelResponse `json:"data"`
	Total   int64                  `json:"total"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
	HasNext bool                   `json:"has_next"`
}
