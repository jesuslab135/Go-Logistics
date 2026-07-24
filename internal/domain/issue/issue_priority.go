package issue

// IssuePriority — port of api/models/issue_model.py:7

type IssuePriority struct {
	ID        int64
	CompanyID int64
	Name      string
	Color     string
	Position  int32
}
