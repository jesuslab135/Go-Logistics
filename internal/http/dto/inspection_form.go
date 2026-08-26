package dto

import (
	"time"
)

type CreateInspectionFormRequest struct {
	// archived_at is not writable: archiving is an action
	// (POST .../archive and .../restore), because a referenced record must be
	// archived rather than deleted and that is a decision, not a field.
	Title            string `json:"title" binding:"omitempty,max=255"`
	Description      string `json:"description"`
	Version          int32  `json:"version"`
	RequireLivePhoto bool   `json:"require_live_photo"`
	AutoCreateIssues *bool  `json:"auto_create_issues"`
	Color            string `json:"color" binding:"omitempty,max=7"`
}

type UpdateInspectionFormRequest struct {
	// archived_at is not writable: archiving is an action
	// (POST .../archive and .../restore), because a referenced record must be
	// archived rather than deleted and that is a decision, not a field.
	Title            string `json:"title" binding:"omitempty,max=255"`
	Description      string `json:"description"`
	Version          int32  `json:"version"`
	RequireLivePhoto bool   `json:"require_live_photo"`
	AutoCreateIssues bool   `json:"auto_create_issues"`
	Color            string `json:"color" binding:"omitempty,max=7"`
}

type InspectionFormResponse struct {
	ID               int64      `json:"id"`
	CompanyID        int64      `json:"company_id"`
	Title            string     `json:"title" binding:"omitempty,max=255"`
	Description      string     `json:"description"`
	Version          int32      `json:"version"`
	RequireLivePhoto bool       `json:"require_live_photo"`
	AutoCreateIssues bool       `json:"auto_create_issues"`
	Color            string     `json:"color" binding:"omitempty,max=7"`
	ArchivedAt       *time.Time `json:"archived_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type InspectionFormPage struct {
	Data    []InspectionFormResponse `json:"data"`
	Total   int64                    `json:"total"`
	Limit   int                      `json:"limit"`
	Offset  int                      `json:"offset"`
	HasNext bool                     `json:"has_next"`
}
