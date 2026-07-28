package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateServiceReminderRequest struct {
	AssetID               int64            `json:"asset_id"`
	ServiceTaskID         *int64           `json:"service_task_id"`
	IsActive              *bool            `json:"is_active"`
	Status                string           `json:"status" binding:"omitempty,max=20"`
	TimeInterval          *int32           `json:"time_interval"`
	TimeFrequency         string           `json:"time_frequency" binding:"omitempty,max=10"`
	NextDueAt             *time.Time       `json:"next_due_at"`
	DueSoonAt             *time.Time       `json:"due_soon_at"`
	DueSoonTimeThreshold  *int32           `json:"due_soon_time_threshold"`
	MeterInterval         *decimal.Decimal `json:"meter_interval"`
	NextDueMeterValue     *decimal.Decimal `json:"next_due_meter_value"`
	DueSoonMeterValue     *decimal.Decimal `json:"due_soon_meter_value"`
	DueSoonMeterThreshold *decimal.Decimal `json:"due_soon_meter_threshold"`
	SnoozeUntil           *time.Time       `json:"snooze_until"`
	LastServiceEntryID    *int64           `json:"last_service_entry_id"`
}

type UpdateServiceReminderRequest struct {
	AssetID               int64            `json:"asset_id"`
	ServiceTaskID         *int64           `json:"service_task_id"`
	IsActive              bool             `json:"is_active"`
	Status                string           `json:"status" binding:"omitempty,max=20"`
	TimeInterval          *int32           `json:"time_interval"`
	TimeFrequency         string           `json:"time_frequency" binding:"omitempty,max=10"`
	NextDueAt             *time.Time       `json:"next_due_at"`
	DueSoonAt             *time.Time       `json:"due_soon_at"`
	DueSoonTimeThreshold  *int32           `json:"due_soon_time_threshold"`
	MeterInterval         *decimal.Decimal `json:"meter_interval"`
	NextDueMeterValue     *decimal.Decimal `json:"next_due_meter_value"`
	DueSoonMeterValue     *decimal.Decimal `json:"due_soon_meter_value"`
	DueSoonMeterThreshold *decimal.Decimal `json:"due_soon_meter_threshold"`
	SnoozeUntil           *time.Time       `json:"snooze_until"`
	LastServiceEntryID    *int64           `json:"last_service_entry_id"`
}

type ServiceReminderResponse struct {
	ID                    int64            `json:"id"`
	CompanyID             int64            `json:"company_id"`
	AssetID               int64            `json:"asset_id"`
	ServiceTaskID         *int64           `json:"service_task_id"`
	IsActive              bool             `json:"is_active"`
	Status                string           `json:"status" binding:"omitempty,max=20"`
	TimeInterval          *int32           `json:"time_interval"`
	TimeFrequency         string           `json:"time_frequency" binding:"omitempty,max=10"`
	NextDueAt             *time.Time       `json:"next_due_at"`
	DueSoonAt             *time.Time       `json:"due_soon_at"`
	DueSoonTimeThreshold  *int32           `json:"due_soon_time_threshold"`
	MeterInterval         *decimal.Decimal `json:"meter_interval"`
	NextDueMeterValue     *decimal.Decimal `json:"next_due_meter_value"`
	DueSoonMeterValue     *decimal.Decimal `json:"due_soon_meter_value"`
	DueSoonMeterThreshold *decimal.Decimal `json:"due_soon_meter_threshold"`
	SnoozeUntil           *time.Time       `json:"snooze_until"`
	LastServiceEntryID    *int64           `json:"last_service_entry_id"`
	CreatedAt             time.Time        `json:"created_at"`
	UpdatedAt             time.Time        `json:"updated_at"`
}

type ServiceReminderPage struct {
	Data    []ServiceReminderResponse `json:"data"`
	Total   int64                     `json:"total"`
	Limit   int                       `json:"limit"`
	Offset  int                       `json:"offset"`
	HasNext bool                      `json:"has_next"`
}
