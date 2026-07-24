package dto

type CreateAxleTemplateRequest struct {
	Name           string `json:"name" binding:"omitempty,max=100"`
	Description    string `json:"description"`
	TotalPositions int32  `json:"total_positions"`
}

type UpdateAxleTemplateRequest struct {
	Name           string `json:"name" binding:"omitempty,max=100"`
	Description    string `json:"description"`
	TotalPositions int32  `json:"total_positions"`
}

type AxleTemplateResponse struct {
	ID             int64  `json:"id"`
	CompanyID      int64  `json:"company_id"`
	Name           string `json:"name" binding:"omitempty,max=100"`
	Description    string `json:"description"`
	TotalPositions int32  `json:"total_positions"`
}

type AxleTemplatePage struct {
	Data    []AxleTemplateResponse `json:"data"`
	Total   int64                  `json:"total"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
	HasNext bool                   `json:"has_next"`
}
