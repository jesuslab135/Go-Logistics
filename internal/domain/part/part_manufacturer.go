package part

import "time"

// PartManufacturer — port of api/models/part_model.py:22

type PartManufacturer struct {
	ID        int64
	CompanyID int64
	Name      string
	Website   string
	CreatedAt time.Time
}
