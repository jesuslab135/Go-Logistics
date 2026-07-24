package tire

import (
	"time"

	"github.com/shopspring/decimal"
)

// TireInspection — port of api/models/tire_model.py:392

type TireInspection struct {
	ID              int64
	TireID          int64
	InspectionDate  time.Time
	Odometer        *int32
	TreadDepth32nds int32
	PSI             *decimal.Decimal
	MeasuredByID    *int64
	Notes           string
}
