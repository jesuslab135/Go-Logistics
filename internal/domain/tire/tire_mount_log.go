package tire

import (
	"time"

	"github.com/shopspring/decimal"
)

// TireMountLog — port of api/models/tire_model.py:340
// Append-only audit log: never updated, never deleted.

type TireMountLog struct {
	ID              int64
	TireID          int64
	VehicleID       int64
	PositionCode    string
	EventType       string
	EventDate       time.Time
	Odometer        *int32
	TreadDepth32nds *int32
	PSI             *decimal.Decimal
	PerformedByID   *int64
	Reason          string
}
