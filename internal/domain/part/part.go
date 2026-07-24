package part

import (
	"time"

	"github.com/shopspring/decimal"
)

// Part — port of api/models/part_model.py:56

type Part struct {
	ID                     int64
	CompanyID              int64
	PartNumber             string
	Description            string
	UsefulLifeMonths       *int32
	UsefulLifeDistance     *int32
	PartCategoryID         *int64
	PartManufacturerID     *int64
	MeasurementUnitID      *int64
	ManufacturerPartNumber string
	SupplierPartNumber     string
	UPC                    string
	UnitCost               *decimal.Decimal
	InventoryItem          bool
	ArchivedAt             *time.Time
	CustomFields           map[string]any
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
