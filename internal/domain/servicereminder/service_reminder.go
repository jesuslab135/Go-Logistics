package servicereminder

import (
	"time"

	"github.com/shopspring/decimal"
)

// ServiceReminder — port of api/models/service_reminder_model.py:4
// Supports time-based intervals, meter-based intervals, or both.

type ServiceReminder struct {
	ID                    int64
	CompanyID             int64
	AssetID               int64
	ServiceTaskID         *int64
	IsActive              bool
	Status                string
	TimeInterval          *int32
	TimeFrequency         string
	NextDueAt             *time.Time
	DueSoonAt             *time.Time
	DueSoonTimeThreshold  *int32
	MeterInterval         *decimal.Decimal
	NextDueMeterValue     *decimal.Decimal
	DueSoonMeterValue     *decimal.Decimal
	DueSoonMeterThreshold *decimal.Decimal
	SnoozeUntil           *time.Time
	LastServiceEntryID    *int64
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
