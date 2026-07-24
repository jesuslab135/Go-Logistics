package document

import "time"

// Warranty — port of api/models/document_model.py:4
// ProviderID references Vendor. Covers either an Asset or a Part; nothing in
// the schema enforces that exactly one is set.

type Warranty struct {
	ID         int64
	CompanyID  int64
	ProviderID int64
	AssetID    *int64
	PartID     *int64
	StartDate  time.Time
	EndDate    time.Time
	Terms      string
	IsActive   bool
}
