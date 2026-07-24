package dto

type CreateAxleDefinitionRequest struct {
	PositionIndex    int32  `json:"position_index"`
	Label            string `json:"label" binding:"omitempty,max=50"`
	AxleRole         string `json:"axle_role" binding:"omitempty,max=10"`
	PositionsPerSide int32  `json:"positions_per_side"`
}

type UpdateAxleDefinitionRequest struct {
	PositionIndex    int32  `json:"position_index"`
	Label            string `json:"label" binding:"omitempty,max=50"`
	AxleRole         string `json:"axle_role" binding:"omitempty,max=10"`
	PositionsPerSide int32  `json:"positions_per_side"`
}

type AxleDefinitionResponse struct {
	ID               int64  `json:"id"`
	TemplateID       int64  `json:"template_id"`
	PositionIndex    int32  `json:"position_index"`
	Label            string `json:"label" binding:"omitempty,max=50"`
	AxleRole         string `json:"axle_role" binding:"omitempty,max=10"`
	PositionsPerSide int32  `json:"positions_per_side"`
}

type AxleDefinitionPage struct {
	Data    []AxleDefinitionResponse `json:"data"`
	Total   int64                    `json:"total"`
	Limit   int                      `json:"limit"`
	Offset  int                      `json:"offset"`
	HasNext bool                     `json:"has_next"`
}
