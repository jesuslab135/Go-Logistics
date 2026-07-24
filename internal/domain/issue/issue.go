package issue

import (
	"time"

	"github.com/shopspring/decimal"
)

// Issue — port of api/models/issue_model.py:124
// ResolvableType/ResolvableID are hand-rolled polymorphism pointing at either
// WorkOrder or ServiceEntry; no FK constraint exists, so IDs can dangle.
// M2M: assigned_to -> api_issue_assigned_to, watchers -> api_issue_watchers.

type Issue struct {
	ID                     int64
	CompanyID              int64
	Number                 string
	AssetID                int64
	AssetType              string
	Name                   string
	Summary                string
	Description            string
	State                  string
	PriorityID             *int64
	FaultID                *int64
	SourceType             string
	InspectionSubmissionID *int64
	ReportedAt             time.Time
	ReportedByID           int64
	DueDate                *time.Time
	DueMeterValue          *decimal.Decimal
	DueSecondaryMeterValue *decimal.Decimal
	Overdue                bool
	ResolvedAt             *time.Time
	ResolvedByID           *int64
	ResolutionNote         string
	ReopenedAt             *time.Time
	ReopenedByID           *int64
	ResolvableType         string
	ResolvableID           *int32
	ClosedAt               *time.Time
	ClosedByID             *int64
	ClosedNote             string
	ExternalID             string
	CreatedByWorkflow      bool
	CommentsCount          int32
	ImagesCount            int32
	DocumentsCount         int32
	Labels                 []string
	CustomFields           map[string]any
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
