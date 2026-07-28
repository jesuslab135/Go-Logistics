package dto

import "time"

type CreateGroupRequest struct {
	Name     string `json:"name" binding:"required,max=100"`
	ParentID *int64 `json:"parent_id" binding:"omitempty,min=1"`
	// ancestry is not accepted: it is derived from parent_id (see group.sql).
	IsDefault bool `json:"is_default"`
}

type UpdateGroupRequest struct {
	Name     string `json:"name" binding:"required,max=100"`
	ParentID *int64 `json:"parent_id" binding:"omitempty,min=1"`
	// ancestry is not accepted: it is derived from parent_id (see group.sql).
	IsDefault bool `json:"is_default"`
}

type GroupResponse struct {
	ID        int64     `json:"id"`
	CompanyID int64     `json:"company_id"`
	Name      string    `json:"name"`
	ParentID  *int64    `json:"parent_id"`
	Ancestry  string    `json:"ancestry"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GroupPage struct {
	Data    []GroupResponse `json:"data"`
	Total   int64           `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	HasNext bool            `json:"has_next"`
}
