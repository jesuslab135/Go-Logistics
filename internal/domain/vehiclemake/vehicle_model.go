package vehiclemake

import "time"

// VehicleModel — port of api/models/vehicle_make_model.py:20

type VehicleModel struct {
	ID        int64
	CompanyID int64
	Name      string
	MakeID    *int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
