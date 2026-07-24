package dto

import "time"

type CreateCompanyRequest struct {
	Name                string  `json:"name" binding:"required,max=200"`
	TaxID               string  `json:"tax_id" binding:"required,max=50"`
	Address             string  `json:"address"`
	Phone               string  `json:"phone" binding:"max=20"`
	Email               string  `json:"email" binding:"omitempty,email,max=254"`
	Website             string  `json:"website" binding:"max=200"`
	Logo                *string `json:"logo" binding:"omitempty,max=100"`
	City                string  `json:"city" binding:"max=100"`
	Region              string  `json:"region" binding:"max=50"`
	PostalCode          string  `json:"postal_code" binding:"max=20"`
	Country             string  `json:"country" binding:"max=50"`
	Timezone            string  `json:"timezone" binding:"max=50"`
	Currency            string  `json:"currency" binding:"max=3"`
	SystemOfMeasurement string  `json:"system_of_measurement" binding:"omitempty,oneof=metric imperial"`
}

type UpdateCompanyRequest struct {
	Name                string  `json:"name" binding:"required,max=200"`
	TaxID               string  `json:"tax_id" binding:"required,max=50"`
	Address             string  `json:"address"`
	Phone               string  `json:"phone" binding:"max=20"`
	Email               string  `json:"email" binding:"omitempty,email,max=254"`
	Website             string  `json:"website" binding:"max=200"`
	Logo                *string `json:"logo" binding:"omitempty,max=100"`
	City                string  `json:"city" binding:"max=100"`
	Region              string  `json:"region" binding:"max=50"`
	PostalCode          string  `json:"postal_code" binding:"max=20"`
	Country             string  `json:"country" binding:"max=50"`
	Timezone            string  `json:"timezone" binding:"max=50"`
	Currency            string  `json:"currency" binding:"max=3"`
	SystemOfMeasurement string  `json:"system_of_measurement" binding:"omitempty,oneof=metric imperial"`
}

type CompanyResponse struct {
	ID                  int64     `json:"id"`
	Name                string    `json:"name"`
	TaxID               string    `json:"tax_id"`
	Address             string    `json:"address"`
	Phone               string    `json:"phone"`
	Email               string    `json:"email"`
	Website             string    `json:"website"`
	Logo                *string   `json:"logo"`
	City                string    `json:"city"`
	Region              string    `json:"region"`
	PostalCode          string    `json:"postal_code"`
	Country             string    `json:"country"`
	Timezone            string    `json:"timezone"`
	Currency            string    `json:"currency"`
	SystemOfMeasurement string    `json:"system_of_measurement"`
	CreatedAt           time.Time `json:"created_at"`
}

// CompanyPage is the paginated companies response. It mirrors
// paginate.Page[CompanyResponse]; declared concretely so the OpenAPI schema
// stays generics-free.
type CompanyPage struct {
	Data    []CompanyResponse `json:"data"`
	Total   int64             `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	HasNext bool              `json:"has_next"`
}
