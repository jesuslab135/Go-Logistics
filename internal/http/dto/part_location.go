package dto

import (
	"time"
)

type CreatePartLocationRequest struct {
	Name       string `json:"name" binding:"omitempty,max=100"`
	Address    string `json:"address" binding:"omitempty,max=200"`
	City       string `json:"city" binding:"omitempty,max=100"`
	Region     string `json:"region" binding:"omitempty,max=50"`
	LocationID *int64 `json:"location_id"`
}

type UpdatePartLocationRequest struct {
	Name       string `json:"name" binding:"omitempty,max=100"`
	Address    string `json:"address" binding:"omitempty,max=200"`
	City       string `json:"city" binding:"omitempty,max=100"`
	Region     string `json:"region" binding:"omitempty,max=50"`
	LocationID *int64 `json:"location_id"`
}

type PartLocationResponse struct {
	ID         int64     `json:"id"`
	CompanyID  int64     `json:"company_id"`
	Name       string    `json:"name" binding:"omitempty,max=100"`
	Address    string    `json:"address" binding:"omitempty,max=200"`
	City       string    `json:"city" binding:"omitempty,max=100"`
	Region     string    `json:"region" binding:"omitempty,max=50"`
	LocationID *int64    `json:"location_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type PartLocationPage struct {
	Data    []PartLocationResponse `json:"data"`
	Total   int64                  `json:"total"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
	HasNext bool                   `json:"has_next"`
}
