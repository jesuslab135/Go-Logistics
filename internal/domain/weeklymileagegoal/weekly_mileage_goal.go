package weeklymileagegoal

import (
	"time"

	"github.com/shopspring/decimal"
)

// WeeklyMileageGoal — port of api/models/weekly_mileage_goal_model.py:4
// total_miles_per_week, sales_opportunity and missing_sales were computed in
// the serializer, not stored.

type WeeklyMileageGoal struct {
	ID                 int64
	CompanyID          int64
	ServiceType        string
	RatePerMile        decimal.Decimal
	WeeklyMileageGoal  int32
	MPGGoal            *decimal.Decimal
	UnitsPerService    int32
	MachinesInWorkshop int32
	MissingMiles       int32
	SortOrder          int32
	IsActive           bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
