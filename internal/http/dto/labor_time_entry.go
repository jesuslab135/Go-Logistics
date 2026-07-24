package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateLaborTimeEntryRequest struct {
	TechnicianID      int64            `json:"technician_id"`
	StartedAt         time.Time        `json:"started_at"`
	EndedAt           *time.Time       `json:"ended_at"`
	DurationSeconds   *int32           `json:"duration_seconds"`
	IsActive          bool             `json:"is_active"`
	ClockInLatitude   *decimal.Decimal `json:"clock_in_latitude"`
	ClockInLongitude  *decimal.Decimal `json:"clock_in_longitude"`
	ClockOutLatitude  *decimal.Decimal `json:"clock_out_latitude"`
	ClockOutLongitude *decimal.Decimal `json:"clock_out_longitude"`
}

type UpdateLaborTimeEntryRequest struct {
	TechnicianID      int64            `json:"technician_id"`
	StartedAt         time.Time        `json:"started_at"`
	EndedAt           *time.Time       `json:"ended_at"`
	DurationSeconds   *int32           `json:"duration_seconds"`
	IsActive          bool             `json:"is_active"`
	ClockInLatitude   *decimal.Decimal `json:"clock_in_latitude"`
	ClockInLongitude  *decimal.Decimal `json:"clock_in_longitude"`
	ClockOutLatitude  *decimal.Decimal `json:"clock_out_latitude"`
	ClockOutLongitude *decimal.Decimal `json:"clock_out_longitude"`
}

type LaborTimeEntryResponse struct {
	ID                int64            `json:"id"`
	SubLineItemID     int64            `json:"sub_line_item_id"`
	TechnicianID      int64            `json:"technician_id"`
	StartedAt         time.Time        `json:"started_at"`
	EndedAt           *time.Time       `json:"ended_at"`
	DurationSeconds   *int32           `json:"duration_seconds"`
	IsActive          bool             `json:"is_active"`
	ClockInLatitude   *decimal.Decimal `json:"clock_in_latitude"`
	ClockInLongitude  *decimal.Decimal `json:"clock_in_longitude"`
	ClockOutLatitude  *decimal.Decimal `json:"clock_out_latitude"`
	ClockOutLongitude *decimal.Decimal `json:"clock_out_longitude"`
	CreatedAt         time.Time        `json:"created_at"`
}

type LaborTimeEntryPage struct {
	Data    []LaborTimeEntryResponse `json:"data"`
	Total   int64                    `json:"total"`
	Limit   int                      `json:"limit"`
	Offset  int                      `json:"offset"`
	HasNext bool                     `json:"has_next"`
}
