package part

import (
	"time"

	"github.com/shopspring/decimal"
)

// PartLocationDetail — port of api/models/part_model.py:92
// The table is part_inventory, not part_location_detail — db_table
// override at :118. Pin that name explicitly in every sqlc query.
// LocationID points at PartLocation, not workorder.Location.

type PartLocationDetail struct {
	ID                         int64
	PartID                     int64
	LocationID                 int64
	AvailableQuantity          decimal.Decimal
	ExpiryDate                 *time.Time
	Aisle                      string
	Row                        string
	Bin                        string
	ReorderPoint               *int32
	ReorderPointEnabled        bool
	ReorderQuantity            *int32
	ReorderPointLeadTimeDays   *int32
	Active                     bool
	TrackInventory             bool
	AverageUnitCost            *decimal.Decimal
	AvailableQuantityUpdatedAt *time.Time
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}
