package organization

import (
	"time"

	"github.com/shopspring/decimal"
)

// Employee — port of api/models/organization_model.py:45
// UserID references Django auth_user. Companies is M2M via api_employee_companies.

type Employee struct {
	ID                   int64
	UserID               *int64
	DefaultCompanyID     *int64
	FirstName            string
	LastName             string
	EmployeeID           string
	RoleID               *int64
	IsActive             bool
	Email                string
	MobilePhone          string
	WorkPhone            string
	JobTitle             string
	StartDate            *time.Time
	LeaveDate            *time.Time
	BirthDate            *time.Time
	HourlyLaborRate      *decimal.Decimal
	IsTechnician         bool
	IsVehicleOperator    bool
	IsAccountOwner       bool
	LicenseClass         string
	LicenseNumber        string
	LicenseState         string
	LicenseExpiry        *time.Time
	StreetAddress        string
	City                 string
	Region               string
	PostalCode           string
	Country              string
	GroupID              *int64
	CustomFields         map[string]any
	TablePreferences     map[string]any
	DashboardPreferences map[string]any
	UpdatedAt            time.Time
}
