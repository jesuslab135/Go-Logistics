package dto

type CreateLocationRequest struct {
	Name     string `json:"name" binding:"omitempty,max=100"`
	IsActive bool   `json:"is_active"`
}

type UpdateLocationRequest struct {
	Name     string `json:"name" binding:"omitempty,max=100"`
	IsActive bool   `json:"is_active"`
}

type LocationResponse struct {
	ID        int64  `json:"id"`
	CompanyID int64  `json:"company_id"`
	Name      string `json:"name" binding:"omitempty,max=100"`
	IsActive  bool   `json:"is_active"`
}

type LocationPage struct {
	Data    []LocationResponse `json:"data"`
	Total   int64              `json:"total"`
	Limit   int                `json:"limit"`
	Offset  int                `json:"offset"`
	HasNext bool               `json:"has_next"`
}
