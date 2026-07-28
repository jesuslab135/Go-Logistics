package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateFuelEntryRequest struct {
	EmployeeID     int64            `json:"employee_id"`
	Date           time.Time        `json:"date"`
	FuelType       string           `json:"fuel_type" binding:"omitempty,max=20"`
	Quantity       decimal.Decimal  `json:"quantity"`
	UnitCost       decimal.Decimal  `json:"unit_cost"`
	TotalCost      decimal.Decimal  `json:"total_cost"`
	Odometer       decimal.Decimal  `json:"odometer"`
	VendorID       int64            `json:"vendor_id"`
	FullTank       *bool            `json:"full_tank"`
	MilesTraveled  *decimal.Decimal `json:"miles_traveled"`
	FuelEfficiency *decimal.Decimal `json:"fuel_efficiency"`
	State          string           `json:"state" binding:"omitempty,max=50"`
	Reference      string           `json:"reference" binding:"omitempty,max=100"`
	Personal       bool             `json:"personal"`
	Reset          bool             `json:"reset"`
	Latitude       *decimal.Decimal `json:"latitude"`
	Longitude      *decimal.Decimal `json:"longitude"`
	ExternalID     string           `json:"external_id" binding:"omitempty,max=100"`
	NoSemana       *string          `json:"no_semana" binding:"omitempty,max=50"`
	EstadoProv     *string          `json:"estado_prov" binding:"omitempty,max=100"`
	OperatorName   *string          `json:"operator_name" binding:"omitempty,max=200"`
}

type UpdateFuelEntryRequest struct {
	EmployeeID     int64            `json:"employee_id"`
	Date           time.Time        `json:"date"`
	FuelType       string           `json:"fuel_type" binding:"omitempty,max=20"`
	Quantity       decimal.Decimal  `json:"quantity"`
	UnitCost       decimal.Decimal  `json:"unit_cost"`
	TotalCost      decimal.Decimal  `json:"total_cost"`
	Odometer       decimal.Decimal  `json:"odometer"`
	VendorID       int64            `json:"vendor_id"`
	FullTank       bool             `json:"full_tank"`
	MilesTraveled  *decimal.Decimal `json:"miles_traveled"`
	FuelEfficiency *decimal.Decimal `json:"fuel_efficiency"`
	State          string           `json:"state" binding:"omitempty,max=50"`
	Reference      string           `json:"reference" binding:"omitempty,max=100"`
	Personal       bool             `json:"personal"`
	Reset          bool             `json:"reset"`
	Latitude       *decimal.Decimal `json:"latitude"`
	Longitude      *decimal.Decimal `json:"longitude"`
	ExternalID     string           `json:"external_id" binding:"omitempty,max=100"`
	NoSemana       *string          `json:"no_semana" binding:"omitempty,max=50"`
	EstadoProv     *string          `json:"estado_prov" binding:"omitempty,max=100"`
	OperatorName   *string          `json:"operator_name" binding:"omitempty,max=200"`
}

type FuelEntryResponse struct {
	ID             int64            `json:"id"`
	AssetID        int64            `json:"asset_id"`
	EmployeeID     int64            `json:"employee_id"`
	Date           time.Time        `json:"date"`
	FuelType       string           `json:"fuel_type" binding:"omitempty,max=20"`
	Quantity       decimal.Decimal  `json:"quantity"`
	UnitCost       decimal.Decimal  `json:"unit_cost"`
	TotalCost      decimal.Decimal  `json:"total_cost"`
	Odometer       decimal.Decimal  `json:"odometer"`
	VendorID       int64            `json:"vendor_id"`
	FullTank       bool             `json:"full_tank"`
	MilesTraveled  *decimal.Decimal `json:"miles_traveled"`
	FuelEfficiency *decimal.Decimal `json:"fuel_efficiency"`
	State          string           `json:"state" binding:"omitempty,max=50"`
	Reference      string           `json:"reference" binding:"omitempty,max=100"`
	Personal       bool             `json:"personal"`
	Reset          bool             `json:"reset"`
	Latitude       *decimal.Decimal `json:"latitude"`
	Longitude      *decimal.Decimal `json:"longitude"`
	ExternalID     string           `json:"external_id" binding:"omitempty,max=100"`
	UpdatedAt      time.Time        `json:"updated_at"`
	NoSemana       *string          `json:"no_semana" binding:"omitempty,max=50"`
	EstadoProv     *string          `json:"estado_prov" binding:"omitempty,max=100"`
	OperatorName   *string          `json:"operator_name" binding:"omitempty,max=200"`
}

type FuelEntryPage struct {
	Data    []FuelEntryResponse `json:"data"`
	Total   int64               `json:"total"`
	Limit   int                 `json:"limit"`
	Offset  int                 `json:"offset"`
	HasNext bool                `json:"has_next"`
}
