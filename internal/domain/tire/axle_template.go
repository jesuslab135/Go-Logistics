package tire

// AxleTemplate — port of api/models/tire_model.py:10
// TotalPositions is denormalized: sum(positions_per_side) * 2, recalculated
// whenever the axle definitions change.

type AxleTemplate struct {
	ID             int64
	CompanyID      int64
	Name           string
	Description    string
	TotalPositions int32
}
