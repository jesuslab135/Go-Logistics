package inspection

import "github.com/shopspring/decimal"

// InspectionSubmissionItem — port of api/models/inspection_model.py:200

type InspectionSubmissionItem struct {
	ID               int64
	SubmissionID     int64
	FormItemID       int64
	ResultStatus     string
	ResultValue      map[string]any
	Remark           string
	Photo            *string
	Latitude         *decimal.Decimal
	Longitude        *decimal.Decimal
	GeneratedIssueID *int64
}
