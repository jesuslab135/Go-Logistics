package fuel

import "time"

// FuelType — port of api/models/fuel_type_model.py:4
// Catalog table; FuelEntry.FuelType is still a CharField, not an FK to this.

type FuelType struct {
	ID        int64
	CompanyID int64
	Name      string
	CreatedAt time.Time
}
