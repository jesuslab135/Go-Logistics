package dto

import (
	"time"
)

type CreatePartCategoryRequest struct {
	Name        string `json:"name" binding:"omitempty,max=100"`
	Description string `json:"description"`
}

type UpdatePartCategoryRequest struct {
	Name        string `json:"name" binding:"omitempty,max=100"`
	Description string `json:"description"`
}

type PartCategoryResponse struct {
	ID          int64     `json:"id"`
	CompanyID   int64     `json:"company_id"`
	Name        string    `json:"name" binding:"omitempty,max=100"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type PartCategoryPage struct {
	Data    []PartCategoryResponse `json:"data"`
	Total   int64                  `json:"total"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
	HasNext bool                   `json:"has_next"`
}
