package dto

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type CreateInspectionSubmissionItemRequest struct {
	FormItemID       int64            `json:"form_item_id"`
	ResultStatus     string           `json:"result_status" binding:"omitempty,max=10"`
	ResultValue      json.RawMessage  `json:"result_value"`
	Remark           string           `json:"remark"`
	Photo            *string          `json:"photo" binding:"omitempty,max=100"`
	Latitude         *decimal.Decimal `json:"latitude"`
	Longitude        *decimal.Decimal `json:"longitude"`
	GeneratedIssueID *int64           `json:"generated_issue_id"`
}

type UpdateInspectionSubmissionItemRequest struct {
	FormItemID       int64            `json:"form_item_id"`
	ResultStatus     string           `json:"result_status" binding:"omitempty,max=10"`
	ResultValue      json.RawMessage  `json:"result_value"`
	Remark           string           `json:"remark"`
	Photo            *string          `json:"photo" binding:"omitempty,max=100"`
	Latitude         *decimal.Decimal `json:"latitude"`
	Longitude        *decimal.Decimal `json:"longitude"`
	GeneratedIssueID *int64           `json:"generated_issue_id"`
}

type InspectionSubmissionItemResponse struct {
	ID               int64            `json:"id"`
	SubmissionID     int64            `json:"submission_id"`
	FormItemID       int64            `json:"form_item_id"`
	ResultStatus     string           `json:"result_status" binding:"omitempty,max=10"`
	ResultValue      json.RawMessage  `json:"result_value"`
	Remark           string           `json:"remark"`
	Photo            *string          `json:"photo" binding:"omitempty,max=100"`
	Latitude         *decimal.Decimal `json:"latitude"`
	Longitude        *decimal.Decimal `json:"longitude"`
	GeneratedIssueID *int64           `json:"generated_issue_id"`
}

type InspectionSubmissionItemPage struct {
	Data    []InspectionSubmissionItemResponse `json:"data"`
	Total   int64                              `json:"total"`
	Limit   int                                `json:"limit"`
	Offset  int                                `json:"offset"`
	HasNext bool                               `json:"has_next"`
}
