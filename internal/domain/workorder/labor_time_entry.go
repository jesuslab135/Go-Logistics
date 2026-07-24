package workorder

import (
	"time"

	"github.com/shopspring/decimal"
)

// LaborTimeEntry — port of api/models/work_order_model.py:399

type LaborTimeEntry struct {
	ID                int64
	SubLineItemID     int64
	TechnicianID      int64
	StartedAt         time.Time
	EndedAt           *time.Time
	DurationSeconds   *int32
	IsActive          bool
	ClockInLatitude   *decimal.Decimal
	ClockInLongitude  *decimal.Decimal
	ClockOutLatitude  *decimal.Decimal
	ClockOutLongitude *decimal.Decimal
	CreatedAt         time.Time
}
