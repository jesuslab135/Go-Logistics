package part

import "time"

// MeasurementUnit — port of api/models/part_model.py:39

type MeasurementUnit struct {
	ID           int64
	CompanyID    int64
	Name         string
	Abbreviation string
	CreatedAt    time.Time
}
