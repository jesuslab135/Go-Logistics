package tire

import (
	"time"

	"github.com/shopspring/decimal"
)

// Tire — port of api/models/tire_model.py:189
// CurrentVehicleID / CurrentPositionCode are denormalized for fast lookups;
// TireInstallation is the authoritative record.

type Tire struct {
	ID                       int64
	CompanyID                int64
	TireIdentificationNumber string
	TireModelID              *int64
	Status                   string
	CurrentTreadDepth32nds   *int32
	CurrentPSI               *decimal.Decimal
	TotalMiles               int32
	CurrentVehicleID         *int64
	CurrentPositionCode      string
	PurchaseDate             *time.Time
	PurchaseCost             *decimal.Decimal
	VendorID                 *int64
	CreatedAt                time.Time
}
