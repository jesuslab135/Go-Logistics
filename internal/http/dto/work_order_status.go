package dto

type CreateWorkOrderStatusRequest struct {
	Name             string `json:"name" binding:"omitempty,max=100"`
	Description      string `json:"description"`
	Color            string `json:"color" binding:"omitempty,max=7"`
	IsDefault        bool   `json:"is_default"`
	MarksAsCompleted bool   `json:"marks_as_completed"`
	Position         int32  `json:"position"`
}

type UpdateWorkOrderStatusRequest struct {
	Name             string `json:"name" binding:"omitempty,max=100"`
	Description      string `json:"description"`
	Color            string `json:"color" binding:"omitempty,max=7"`
	IsDefault        bool   `json:"is_default"`
	MarksAsCompleted bool   `json:"marks_as_completed"`
	Position         int32  `json:"position"`
}

type WorkOrderStatusResponse struct {
	ID               int64  `json:"id"`
	CompanyID        int64  `json:"company_id"`
	Name             string `json:"name" binding:"omitempty,max=100"`
	Description      string `json:"description"`
	Color            string `json:"color" binding:"omitempty,max=7"`
	IsDefault        bool   `json:"is_default"`
	MarksAsCompleted bool   `json:"marks_as_completed"`
	Position         int32  `json:"position"`
}

type WorkOrderStatusPage struct {
	Data    []WorkOrderStatusResponse `json:"data"`
	Total   int64                     `json:"total"`
	Limit   int                       `json:"limit"`
	Offset  int                       `json:"offset"`
	HasNext bool                      `json:"has_next"`
}
