package tire

// VehicleAxleConfig — port of api/models/tire_model.py:434
// Shared primary key on Asset, like Vehicle and Trailer: no own id column.

type VehicleAxleConfig struct {
	VehicleID   int64
	TemplateID  int64
	DisplayName string
}
