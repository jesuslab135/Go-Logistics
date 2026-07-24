package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type CreateIssueRequest struct {
	Number                 string           `json:"number" binding:"omitempty,max=50"`
	AssetID                int64            `json:"asset_id"`
	AssetType              string           `json:"asset_type" binding:"omitempty,max=20"`
	Name                   string           `json:"name" binding:"omitempty,max=255"`
	Summary                string           `json:"summary" binding:"omitempty,max=255"`
	Description            string           `json:"description"`
	State                  string           `json:"state" binding:"omitempty,max=20"`
	PriorityID             *int64           `json:"priority_id"`
	FaultID                *int64           `json:"fault_id"`
	SourceType             string           `json:"source_type" binding:"omitempty,max=20"`
	InspectionSubmissionID *int64           `json:"inspection_submission_id"`
	ReportedAt             time.Time        `json:"reported_at"`
	ReportedByID           int64            `json:"reported_by_id"`
	DueDate                *time.Time       `json:"due_date"`
	DueMeterValue          *decimal.Decimal `json:"due_meter_value"`
	DueSecondaryMeterValue *decimal.Decimal `json:"due_secondary_meter_value"`
	Overdue                bool             `json:"overdue"`
	ResolvedAt             *time.Time       `json:"resolved_at"`
	ResolvedByID           *int64           `json:"resolved_by_id"`
	ResolutionNote         string           `json:"resolution_note"`
	ReopenedAt             *time.Time       `json:"reopened_at"`
	ReopenedByID           *int64           `json:"reopened_by_id"`
	ResolvableType         string           `json:"resolvable_type" binding:"omitempty,max=30"`
	ResolvableID           *int32           `json:"resolvable_id"`
	ClosedAt               *time.Time       `json:"closed_at"`
	ClosedByID             *int64           `json:"closed_by_id"`
	ClosedNote             string           `json:"closed_note"`
	ExternalID             string           `json:"external_id" binding:"omitempty,max=100"`
	CreatedByWorkflow      bool             `json:"created_by_workflow"`
	CommentsCount          int32            `json:"comments_count"`
	ImagesCount            int32            `json:"images_count"`
	DocumentsCount         int32            `json:"documents_count"`
	Labels                 json.RawMessage  `json:"labels"`
	CustomFields           json.RawMessage  `json:"custom_fields"`
}

type UpdateIssueRequest struct {
	Number                 string           `json:"number" binding:"omitempty,max=50"`
	AssetID                int64            `json:"asset_id"`
	AssetType              string           `json:"asset_type" binding:"omitempty,max=20"`
	Name                   string           `json:"name" binding:"omitempty,max=255"`
	Summary                string           `json:"summary" binding:"omitempty,max=255"`
	Description            string           `json:"description"`
	State                  string           `json:"state" binding:"omitempty,max=20"`
	PriorityID             *int64           `json:"priority_id"`
	FaultID                *int64           `json:"fault_id"`
	SourceType             string           `json:"source_type" binding:"omitempty,max=20"`
	InspectionSubmissionID *int64           `json:"inspection_submission_id"`
	ReportedAt             time.Time        `json:"reported_at"`
	ReportedByID           int64            `json:"reported_by_id"`
	DueDate                *time.Time       `json:"due_date"`
	DueMeterValue          *decimal.Decimal `json:"due_meter_value"`
	DueSecondaryMeterValue *decimal.Decimal `json:"due_secondary_meter_value"`
	Overdue                bool             `json:"overdue"`
	ResolvedAt             *time.Time       `json:"resolved_at"`
	ResolvedByID           *int64           `json:"resolved_by_id"`
	ResolutionNote         string           `json:"resolution_note"`
	ReopenedAt             *time.Time       `json:"reopened_at"`
	ReopenedByID           *int64           `json:"reopened_by_id"`
	ResolvableType         string           `json:"resolvable_type" binding:"omitempty,max=30"`
	ResolvableID           *int32           `json:"resolvable_id"`
	ClosedAt               *time.Time       `json:"closed_at"`
	ClosedByID             *int64           `json:"closed_by_id"`
	ClosedNote             string           `json:"closed_note"`
	ExternalID             string           `json:"external_id" binding:"omitempty,max=100"`
	CreatedByWorkflow      bool             `json:"created_by_workflow"`
	CommentsCount          int32            `json:"comments_count"`
	ImagesCount            int32            `json:"images_count"`
	DocumentsCount         int32            `json:"documents_count"`
	Labels                 json.RawMessage  `json:"labels"`
	CustomFields           json.RawMessage  `json:"custom_fields"`
}

type IssueResponse struct {
	ID                     int64            `json:"id"`
	CompanyID              int64            `json:"company_id"`
	Number                 string           `json:"number" binding:"omitempty,max=50"`
	AssetID                int64            `json:"asset_id"`
	AssetType              string           `json:"asset_type" binding:"omitempty,max=20"`
	Name                   string           `json:"name" binding:"omitempty,max=255"`
	Summary                string           `json:"summary" binding:"omitempty,max=255"`
	Description            string           `json:"description"`
	State                  string           `json:"state" binding:"omitempty,max=20"`
	PriorityID             *int64           `json:"priority_id"`
	FaultID                *int64           `json:"fault_id"`
	SourceType             string           `json:"source_type" binding:"omitempty,max=20"`
	InspectionSubmissionID *int64           `json:"inspection_submission_id"`
	ReportedAt             time.Time        `json:"reported_at"`
	ReportedByID           int64            `json:"reported_by_id"`
	DueDate                *time.Time       `json:"due_date"`
	DueMeterValue          *decimal.Decimal `json:"due_meter_value"`
	DueSecondaryMeterValue *decimal.Decimal `json:"due_secondary_meter_value"`
	Overdue                bool             `json:"overdue"`
	ResolvedAt             *time.Time       `json:"resolved_at"`
	ResolvedByID           *int64           `json:"resolved_by_id"`
	ResolutionNote         string           `json:"resolution_note"`
	ReopenedAt             *time.Time       `json:"reopened_at"`
	ReopenedByID           *int64           `json:"reopened_by_id"`
	ResolvableType         string           `json:"resolvable_type" binding:"omitempty,max=30"`
	ResolvableID           *int32           `json:"resolvable_id"`
	ClosedAt               *time.Time       `json:"closed_at"`
	ClosedByID             *int64           `json:"closed_by_id"`
	ClosedNote             string           `json:"closed_note"`
	ExternalID             string           `json:"external_id" binding:"omitempty,max=100"`
	CreatedByWorkflow      bool             `json:"created_by_workflow"`
	CommentsCount          int32            `json:"comments_count"`
	ImagesCount            int32            `json:"images_count"`
	DocumentsCount         int32            `json:"documents_count"`
	Labels                 json.RawMessage  `json:"labels"`
	CustomFields           json.RawMessage  `json:"custom_fields"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
}

type IssuePage struct {
	Data    []IssueResponse `json:"data"`
	Total   int64           `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	HasNext bool            `json:"has_next"`
}
