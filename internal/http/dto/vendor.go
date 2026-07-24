package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type CreateVendorRequest struct {
	Name               string           `json:"name" binding:"omitempty,max=200"`
	IsMobileService    bool             `json:"is_mobile_service"`
	StreetAddress      string           `json:"street_address" binding:"omitempty,max=200"`
	StreetAddressLine2 string           `json:"street_address_line_2" binding:"omitempty,max=200"`
	City               string           `json:"city" binding:"omitempty,max=100"`
	Region             string           `json:"region" binding:"omitempty,max=50"`
	PostalCode         string           `json:"postal_code" binding:"omitempty,max=20"`
	Country            string           `json:"country" binding:"omitempty,max=50"`
	Phone              string           `json:"phone" binding:"omitempty,max=20"`
	Website            string           `json:"website" binding:"omitempty,max=200"`
	ContactName        string           `json:"contact_name" binding:"omitempty,max=200"`
	ContactPhone       string           `json:"contact_phone" binding:"omitempty,max=20"`
	ContactEmail       string           `json:"contact_email" binding:"omitempty,max=254"`
	ExternalID         string           `json:"external_id" binding:"omitempty,max=100"`
	Latitude           *decimal.Decimal `json:"latitude"`
	Longitude          *decimal.Decimal `json:"longitude"`
	IsFuelVendor       bool             `json:"is_fuel_vendor"`
	IsServiceVendor    bool             `json:"is_service_vendor"`
	IsPartsVendor      bool             `json:"is_parts_vendor"`
	Labels             json.RawMessage  `json:"labels"`
	ArchivedAt         *time.Time       `json:"archived_at"`
	CustomFields       json.RawMessage  `json:"custom_fields"`
}

type UpdateVendorRequest struct {
	Name               string           `json:"name" binding:"omitempty,max=200"`
	IsMobileService    bool             `json:"is_mobile_service"`
	StreetAddress      string           `json:"street_address" binding:"omitempty,max=200"`
	StreetAddressLine2 string           `json:"street_address_line_2" binding:"omitempty,max=200"`
	City               string           `json:"city" binding:"omitempty,max=100"`
	Region             string           `json:"region" binding:"omitempty,max=50"`
	PostalCode         string           `json:"postal_code" binding:"omitempty,max=20"`
	Country            string           `json:"country" binding:"omitempty,max=50"`
	Phone              string           `json:"phone" binding:"omitempty,max=20"`
	Website            string           `json:"website" binding:"omitempty,max=200"`
	ContactName        string           `json:"contact_name" binding:"omitempty,max=200"`
	ContactPhone       string           `json:"contact_phone" binding:"omitempty,max=20"`
	ContactEmail       string           `json:"contact_email" binding:"omitempty,max=254"`
	ExternalID         string           `json:"external_id" binding:"omitempty,max=100"`
	Latitude           *decimal.Decimal `json:"latitude"`
	Longitude          *decimal.Decimal `json:"longitude"`
	IsFuelVendor       bool             `json:"is_fuel_vendor"`
	IsServiceVendor    bool             `json:"is_service_vendor"`
	IsPartsVendor      bool             `json:"is_parts_vendor"`
	Labels             json.RawMessage  `json:"labels"`
	ArchivedAt         *time.Time       `json:"archived_at"`
	CustomFields       json.RawMessage  `json:"custom_fields"`
}

type VendorResponse struct {
	ID                 int64            `json:"id"`
	CompanyID          int64            `json:"company_id"`
	Name               string           `json:"name" binding:"omitempty,max=200"`
	IsMobileService    bool             `json:"is_mobile_service"`
	StreetAddress      string           `json:"street_address" binding:"omitempty,max=200"`
	StreetAddressLine2 string           `json:"street_address_line_2" binding:"omitempty,max=200"`
	City               string           `json:"city" binding:"omitempty,max=100"`
	Region             string           `json:"region" binding:"omitempty,max=50"`
	PostalCode         string           `json:"postal_code" binding:"omitempty,max=20"`
	Country            string           `json:"country" binding:"omitempty,max=50"`
	Phone              string           `json:"phone" binding:"omitempty,max=20"`
	Website            string           `json:"website" binding:"omitempty,max=200"`
	ContactName        string           `json:"contact_name" binding:"omitempty,max=200"`
	ContactPhone       string           `json:"contact_phone" binding:"omitempty,max=20"`
	ContactEmail       string           `json:"contact_email" binding:"omitempty,max=254"`
	ExternalID         string           `json:"external_id" binding:"omitempty,max=100"`
	Latitude           *decimal.Decimal `json:"latitude"`
	Longitude          *decimal.Decimal `json:"longitude"`
	IsFuelVendor       bool             `json:"is_fuel_vendor"`
	IsServiceVendor    bool             `json:"is_service_vendor"`
	IsPartsVendor      bool             `json:"is_parts_vendor"`
	Labels             json.RawMessage  `json:"labels"`
	ArchivedAt         *time.Time       `json:"archived_at"`
	CustomFields       json.RawMessage  `json:"custom_fields"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

type VendorPage struct {
	Data    []VendorResponse `json:"data"`
	Total   int64            `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
	HasNext bool             `json:"has_next"`
}
