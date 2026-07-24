package dto

type CreateWheelPositionDefinitionRequest struct {
	Code string `json:"code" binding:"omitempty,max=10"`
	Side string `json:"side" binding:"omitempty,max=1"`
	Slot int32  `json:"slot"`
}

type UpdateWheelPositionDefinitionRequest struct {
	Code string `json:"code" binding:"omitempty,max=10"`
	Side string `json:"side" binding:"omitempty,max=1"`
	Slot int32  `json:"slot"`
}

type WheelPositionDefinitionResponse struct {
	ID     int64  `json:"id"`
	AxleID int64  `json:"axle_id"`
	Code   string `json:"code" binding:"omitempty,max=10"`
	Side   string `json:"side" binding:"omitempty,max=1"`
	Slot   int32  `json:"slot"`
}

type WheelPositionDefinitionPage struct {
	Data    []WheelPositionDefinitionResponse `json:"data"`
	Total   int64                             `json:"total"`
	Limit   int                               `json:"limit"`
	Offset  int                               `json:"offset"`
	HasNext bool                              `json:"has_next"`
}
