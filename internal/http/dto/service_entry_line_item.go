package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateServiceEntryLineItemRequest struct {
	LineItemType      string          `json:"line_item_type" binding:"omitempty,max=20"`
	Description       string          `json:"description" binding:"omitempty,max=255"`
	ServiceTaskID     *int64          `json:"service_task_id"`
	PartID            *int64          `json:"part_id"`
	TechnicianID      *int64          `json:"technician_id"`
	TireID            *int64          `json:"tire_id"`
	ServiceReminderID *int64          `json:"service_reminder_id"`
	UnitCost          decimal.Decimal `json:"unit_cost"`
	Quantity          decimal.Decimal `json:"quantity"`
	PartsCost         decimal.Decimal `json:"parts_cost"`
	LaborCost         decimal.Decimal `json:"labor_cost"`
	Subtotal          decimal.Decimal `json:"subtotal"`
	Position          int32           `json:"position"`
}

type UpdateServiceEntryLineItemRequest struct {
	LineItemType      string          `json:"line_item_type" binding:"omitempty,max=20"`
	Description       string          `json:"description" binding:"omitempty,max=255"`
	ServiceTaskID     *int64          `json:"service_task_id"`
	PartID            *int64          `json:"part_id"`
	TechnicianID      *int64          `json:"technician_id"`
	TireID            *int64          `json:"tire_id"`
	ServiceReminderID *int64          `json:"service_reminder_id"`
	UnitCost          decimal.Decimal `json:"unit_cost"`
	Quantity          decimal.Decimal `json:"quantity"`
	PartsCost         decimal.Decimal `json:"parts_cost"`
	LaborCost         decimal.Decimal `json:"labor_cost"`
	Subtotal          decimal.Decimal `json:"subtotal"`
	Position          int32           `json:"position"`
}

type ServiceEntryLineItemResponse struct {
	ID                int64           `json:"id"`
	ServiceEntryID    int64           `json:"service_entry_id"`
	LineItemType      string          `json:"line_item_type" binding:"omitempty,max=20"`
	Description       string          `json:"description" binding:"omitempty,max=255"`
	ServiceTaskID     *int64          `json:"service_task_id"`
	PartID            *int64          `json:"part_id"`
	TechnicianID      *int64          `json:"technician_id"`
	TireID            *int64          `json:"tire_id"`
	ServiceReminderID *int64          `json:"service_reminder_id"`
	UnitCost          decimal.Decimal `json:"unit_cost"`
	Quantity          decimal.Decimal `json:"quantity"`
	PartsCost         decimal.Decimal `json:"parts_cost"`
	LaborCost         decimal.Decimal `json:"labor_cost"`
	Subtotal          decimal.Decimal `json:"subtotal"`
	Position          int32           `json:"position"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type ServiceEntryLineItemPage struct {
	Data    []ServiceEntryLineItemResponse `json:"data"`
	Total   int64                          `json:"total"`
	Limit   int                            `json:"limit"`
	Offset  int                            `json:"offset"`
	HasNext bool                           `json:"has_next"`
}
