package fuel

import (
	"time"

	"github.com/shopspring/decimal"
)

// FuelEntry — port of api/models/fuel_model.py:5
// NoSemana, EstadoProv and OperatorName are blank=True AND null=True, so both
// NULL and "" are reachable in existing rows (:51,:58,:65).
// TotalCost, MilesTraveled and FuelEfficiency were computed in save() against
// the previous entry of the same fuel_type; that belongs in a service.

type FuelEntry struct {
	ID             int64
	AssetID        int64
	EmployeeID     int64
	Date           time.Time
	FuelType       string
	Quantity       decimal.Decimal
	UnitCost       decimal.Decimal
	TotalCost      decimal.Decimal
	Odometer       decimal.Decimal
	VendorID       int64
	FullTank       bool
	MilesTraveled  *decimal.Decimal
	FuelEfficiency *decimal.Decimal
	State          string
	Reference      string
	Personal       bool
	Reset          bool
	Latitude       *decimal.Decimal
	Longitude      *decimal.Decimal
	ExternalID     string
	UpdatedAt      time.Time
	NoSemana       *string
	EstadoProv     *string
	OperatorName   *string
}
