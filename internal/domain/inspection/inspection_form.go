package inspection

import "time"

// InspectionForm — port of api/models/inspection_model.py:4

type InspectionForm struct {
	ID               int64
	CompanyID        int64
	Title            string
	Description      string
	Version          int32
	RequireLivePhoto bool
	AutoCreateIssues bool
	Color            string
	ArchivedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (f InspectionForm) IsActive() bool {
	return f.ArchivedAt == nil
}
