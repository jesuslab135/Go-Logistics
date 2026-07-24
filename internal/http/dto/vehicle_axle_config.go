package dto

type UpsertVehicleAxleConfigRequest struct {
	TemplateID  int64  `json:"template_id"`
	DisplayName string `json:"display_name" binding:"omitempty,max=100"`
}

type VehicleAxleConfigResponse struct {
	VehicleID   int64  `json:"vehicle_id"`
	TemplateID  int64  `json:"template_id"`
	DisplayName string `json:"display_name" binding:"omitempty,max=100"`
}
