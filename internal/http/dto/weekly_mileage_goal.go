package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type CreateWeeklyMileageGoalRequest struct {
	ServiceType        string           `json:"service_type" binding:"omitempty,max=100"`
	RatePerMile        decimal.Decimal  `json:"rate_per_mile"`
	WeeklyMileageGoal  int32            `json:"weekly_mileage_goal"`
	MpgGoal            *decimal.Decimal `json:"mpg_goal"`
	UnitsPerService    int32            `json:"units_per_service"`
	MachinesInWorkshop int32            `json:"machines_in_workshop"`
	MissingMiles       int32            `json:"missing_miles"`
	SortOrder          int32            `json:"sort_order"`
	IsActive           bool             `json:"is_active"`
}

type UpdateWeeklyMileageGoalRequest struct {
	ServiceType        string           `json:"service_type" binding:"omitempty,max=100"`
	RatePerMile        decimal.Decimal  `json:"rate_per_mile"`
	WeeklyMileageGoal  int32            `json:"weekly_mileage_goal"`
	MpgGoal            *decimal.Decimal `json:"mpg_goal"`
	UnitsPerService    int32            `json:"units_per_service"`
	MachinesInWorkshop int32            `json:"machines_in_workshop"`
	MissingMiles       int32            `json:"missing_miles"`
	SortOrder          int32            `json:"sort_order"`
	IsActive           bool             `json:"is_active"`
}

type WeeklyMileageGoalResponse struct {
	ID                 int64            `json:"id"`
	CompanyID          int64            `json:"company_id"`
	ServiceType        string           `json:"service_type" binding:"omitempty,max=100"`
	RatePerMile        decimal.Decimal  `json:"rate_per_mile"`
	WeeklyMileageGoal  int32            `json:"weekly_mileage_goal"`
	MpgGoal            *decimal.Decimal `json:"mpg_goal"`
	UnitsPerService    int32            `json:"units_per_service"`
	MachinesInWorkshop int32            `json:"machines_in_workshop"`
	MissingMiles       int32            `json:"missing_miles"`
	SortOrder          int32            `json:"sort_order"`
	IsActive           bool             `json:"is_active"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

type WeeklyMileageGoalPage struct {
	Data    []WeeklyMileageGoalResponse `json:"data"`
	Total   int64                       `json:"total"`
	Limit   int                         `json:"limit"`
	Offset  int                         `json:"offset"`
	HasNext bool                        `json:"has_next"`
}
