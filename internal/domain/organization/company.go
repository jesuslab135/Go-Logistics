package organization

import "time"

// Company — port of api/models/organization_model.py:5

type Company struct {
	ID                  int64
	Name                string
	TaxID               string
	Address             string
	CreatedAt           time.Time
	Phone               string
	Email               string
	Website             string
	Logo                *string
	City                string
	Region              string
	PostalCode          string
	Country             string
	Timezone            string
	Currency            string
	SystemOfMeasurement string
}

type MeasurementSystem string

const (
	Metric   MeasurementSystem = "metric"
	Imperial MeasurementSystem = "imperial"
)

func (m MeasurementSystem) Valid() bool {
	return m == Metric || m == Imperial
}
