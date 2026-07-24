package dto

import (
	"encoding/json"
)

type CreateRoleRequest struct {
	Name        string          `json:"name" binding:"omitempty,max=100"`
	IsAdmin     bool            `json:"is_admin"`
	Permissions json.RawMessage `json:"permissions"`
}

type UpdateRoleRequest struct {
	Name        string          `json:"name" binding:"omitempty,max=100"`
	IsAdmin     bool            `json:"is_admin"`
	Permissions json.RawMessage `json:"permissions"`
}

type RoleResponse struct {
	ID          int64           `json:"id"`
	CompanyID   int64           `json:"company_id"`
	Name        string          `json:"name" binding:"omitempty,max=100"`
	IsAdmin     bool            `json:"is_admin"`
	Permissions json.RawMessage `json:"permissions"`
}

type RolePage struct {
	Data    []RoleResponse `json:"data"`
	Total   int64          `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
	HasNext bool           `json:"has_next"`
}
