package part

import "time"

// PartLocation — port of api/models/part_model.py:127
// LocationID links a warehouse to a workorder.Location (shop), nullable.

type PartLocation struct {
	ID         int64
	CompanyID  int64
	Name       string
	Address    string
	City       string
	Region     string
	LocationID *int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
