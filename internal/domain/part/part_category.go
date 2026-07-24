package part

import "time"

// PartCategory — port of api/models/part_model.py:4

type PartCategory struct {
	ID          int64
	CompanyID   int64
	Name        string
	Description string
	CreatedAt   time.Time
}
