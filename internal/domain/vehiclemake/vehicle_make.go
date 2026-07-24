package vehiclemake

import "time"

// VehicleMake — port of api/models/vehicle_make_model.py:3

type VehicleMake struct {
	ID        int64
	CompanyID int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
