package dto

import (
	"time"
)

type CreateServiceTaskRequest struct {
	Name                    string     `json:"name" binding:"omitempty,max=255"`
	Description             string     `json:"description"`
	ExpectedDurationSeconds *int32     `json:"expected_duration_seconds"`
	ParentTaskID            *int64     `json:"parent_task_id"`
	ArchivedAt              *time.Time `json:"archived_at"`
}

type UpdateServiceTaskRequest struct {
	Name                    string     `json:"name" binding:"omitempty,max=255"`
	Description             string     `json:"description"`
	ExpectedDurationSeconds *int32     `json:"expected_duration_seconds"`
	ParentTaskID            *int64     `json:"parent_task_id"`
	ArchivedAt              *time.Time `json:"archived_at"`
}

type ServiceTaskResponse struct {
	ID                      int64      `json:"id"`
	CompanyID               int64      `json:"company_id"`
	Name                    string     `json:"name" binding:"omitempty,max=255"`
	Description             string     `json:"description"`
	ExpectedDurationSeconds *int32     `json:"expected_duration_seconds"`
	ParentTaskID            *int64     `json:"parent_task_id"`
	ArchivedAt              *time.Time `json:"archived_at"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

type ServiceTaskPage struct {
	Data    []ServiceTaskResponse `json:"data"`
	Total   int64                 `json:"total"`
	Limit   int                   `json:"limit"`
	Offset  int                   `json:"offset"`
	HasNext bool                  `json:"has_next"`
}
