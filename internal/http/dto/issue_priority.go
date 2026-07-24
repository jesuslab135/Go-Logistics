package dto

type CreateIssuePriorityRequest struct {
	Name     string `json:"name" binding:"omitempty,max=50"`
	Color    string `json:"color" binding:"omitempty,max=7"`
	Position int32  `json:"position"`
}

type UpdateIssuePriorityRequest struct {
	Name     string `json:"name" binding:"omitempty,max=50"`
	Color    string `json:"color" binding:"omitempty,max=7"`
	Position int32  `json:"position"`
}

type IssuePriorityResponse struct {
	ID        int64  `json:"id"`
	CompanyID int64  `json:"company_id"`
	Name      string `json:"name" binding:"omitempty,max=50"`
	Color     string `json:"color" binding:"omitempty,max=7"`
	Position  int32  `json:"position"`
}

type IssuePriorityPage struct {
	Data    []IssuePriorityResponse `json:"data"`
	Total   int64                   `json:"total"`
	Limit   int                     `json:"limit"`
	Offset  int                     `json:"offset"`
	HasNext bool                    `json:"has_next"`
}
