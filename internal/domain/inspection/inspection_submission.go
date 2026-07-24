package inspection

import (
	"time"

	"github.com/shopspring/decimal"
)

// InspectionSubmission — port of api/models/inspection_model.py:114

type InspectionSubmission struct {
	ID                 int64
	CompanyID          int64
	FormID             int64
	AssetID            int64
	SubmittedByID      int64
	StartedAt          time.Time
	SubmittedAt        time.Time
	DurationSeconds    *int32
	StartingLatitude   *decimal.Decimal
	StartingLongitude  *decimal.Decimal
	SubmittedLatitude  *decimal.Decimal
	SubmittedLongitude *decimal.Decimal
	Signature          *string
	Odometer           *decimal.Decimal
	TotalItems         int32
	FailedItemsCount   int32
	PassedItemsCount   int32
	CommentsCount      int32
	ImagesCount        int32
	GeneralNotes       string
	CreatedAt          time.Time
}

func (s InspectionSubmission) HasFailures() bool {
	return s.FailedItemsCount > 0
}
