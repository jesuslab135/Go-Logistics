package dto

import (
	"time"
)

type CreateWarrantyRequest struct {
	ProviderID int64     `json:"provider_id"`
	AssetID    *int64    `json:"asset_id"`
	PartID     *int64    `json:"part_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Terms      string    `json:"terms"`
	IsActive   *bool     `json:"is_active"`
}

type UpdateWarrantyRequest struct {
	ProviderID int64     `json:"provider_id"`
	AssetID    *int64    `json:"asset_id"`
	PartID     *int64    `json:"part_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Terms      string    `json:"terms"`
	IsActive   bool      `json:"is_active"`
}

type WarrantyResponse struct {
	ID         int64     `json:"id"`
	CompanyID  int64     `json:"company_id"`
	ProviderID int64     `json:"provider_id"`
	AssetID    *int64    `json:"asset_id"`
	PartID     *int64    `json:"part_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Terms      string    `json:"terms"`
	IsActive   bool      `json:"is_active"`
}

type WarrantyPage struct {
	Data    []WarrantyResponse `json:"data"`
	Total   int64              `json:"total"`
	Limit   int                `json:"limit"`
	Offset  int                `json:"offset"`
	HasNext bool               `json:"has_next"`
}
