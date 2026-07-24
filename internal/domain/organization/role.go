package organization

// Role — port of api/models/organization_model.py:34

type Role struct {
	ID          int64
	CompanyID   int64
	Name        string
	IsAdmin     bool
	Permissions map[string]any
}
