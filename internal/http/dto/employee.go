package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type CreateEmployeeRequest struct {
	UserID           *int64 `json:"user_id"`
	DefaultCompanyID *int64 `json:"default_company_id"`
	FirstName        string `json:"first_name" binding:"omitempty,max=100"`
	LastName         string `json:"last_name" binding:"omitempty,max=100"`
	EmployeeID       string `json:"employee_id" binding:"omitempty,max=50"`
	// IsActive is the new membership's status in the session's company.
	// Omitted means active.
	IsActive          *bool            `json:"is_active"`
	Email             string           `json:"email" binding:"omitempty,max=254"`
	MobilePhone       string           `json:"mobile_phone" binding:"omitempty,max=20"`
	WorkPhone         string           `json:"work_phone" binding:"omitempty,max=20"`
	JobTitle          string           `json:"job_title" binding:"omitempty,max=100"`
	StartDate         *time.Time       `json:"start_date"`
	LeaveDate         *time.Time       `json:"leave_date"`
	BirthDate         *time.Time       `json:"birth_date"`
	HourlyLaborRate   *decimal.Decimal `json:"hourly_labor_rate"`
	IsTechnician      bool             `json:"is_technician"`
	IsVehicleOperator bool             `json:"is_vehicle_operator"`
	// RoleID is the role the new employee gets in the session's company. Setting
	// it requires an administrator of that company, and the role must belong to
	// it. Omitted or null creates a membership with no role, which grants nothing.
	RoleID               OptionalInt64   `json:"role_id" swaggertype:"integer"`
	LicenseClass         string          `json:"license_class" binding:"omitempty,max=10"`
	LicenseNumber        string          `json:"license_number" binding:"omitempty,max=50"`
	LicenseState         string          `json:"license_state" binding:"omitempty,max=50"`
	LicenseExpiry        *time.Time      `json:"license_expiry"`
	StreetAddress        string          `json:"street_address" binding:"omitempty,max=200"`
	City                 string          `json:"city" binding:"omitempty,max=100"`
	Region               string          `json:"region" binding:"omitempty,max=50"`
	PostalCode           string          `json:"postal_code" binding:"omitempty,max=20"`
	Country              string          `json:"country" binding:"omitempty,max=50"`
	GroupID              *int64          `json:"group_id"`
	CustomFields         json.RawMessage `json:"custom_fields" swaggertype:"object"`
	TablePreferences     json.RawMessage `json:"table_preferences"`
	DashboardPreferences json.RawMessage `json:"dashboard_preferences"`
}

type UpdateEmployeeRequest struct {
	UserID           *int64 `json:"user_id"`
	DefaultCompanyID *int64 `json:"default_company_id"`
	FirstName        string `json:"first_name" binding:"omitempty,max=100"`
	LastName         string `json:"last_name" binding:"omitempty,max=100"`
	EmployeeID       string `json:"employee_id" binding:"omitempty,max=50"`
	// IsActive is the employee's status in the session's company only: false
	// suspends them here while they keep working in their other companies.
	// Omitted leaves it unchanged. The account owner and the caller themselves
	// cannot be deactivated.
	IsActive          *bool            `json:"is_active"`
	Email             string           `json:"email" binding:"omitempty,max=254"`
	MobilePhone       string           `json:"mobile_phone" binding:"omitempty,max=20"`
	WorkPhone         string           `json:"work_phone" binding:"omitempty,max=20"`
	JobTitle          string           `json:"job_title" binding:"omitempty,max=100"`
	StartDate         *time.Time       `json:"start_date"`
	LeaveDate         *time.Time       `json:"leave_date"`
	BirthDate         *time.Time       `json:"birth_date"`
	HourlyLaborRate   *decimal.Decimal `json:"hourly_labor_rate"`
	IsTechnician      bool             `json:"is_technician"`
	IsVehicleOperator bool             `json:"is_vehicle_operator"`
	// RoleID changes the employee's role in the session's company. Omitted
	// leaves it unchanged; null removes it. Changing it requires an
	// administrator of that company, the role must belong to it, and an
	// administrator cannot change their own role (the account owner can).
	RoleID               OptionalInt64   `json:"role_id" swaggertype:"integer"`
	LicenseClass         string          `json:"license_class" binding:"omitempty,max=10"`
	LicenseNumber        string          `json:"license_number" binding:"omitempty,max=50"`
	LicenseState         string          `json:"license_state" binding:"omitempty,max=50"`
	LicenseExpiry        *time.Time      `json:"license_expiry"`
	StreetAddress        string          `json:"street_address" binding:"omitempty,max=200"`
	City                 string          `json:"city" binding:"omitempty,max=100"`
	Region               string          `json:"region" binding:"omitempty,max=50"`
	PostalCode           string          `json:"postal_code" binding:"omitempty,max=20"`
	Country              string          `json:"country" binding:"omitempty,max=50"`
	GroupID              *int64          `json:"group_id"`
	CustomFields         json.RawMessage `json:"custom_fields" swaggertype:"object"`
	TablePreferences     json.RawMessage `json:"table_preferences"`
	DashboardPreferences json.RawMessage `json:"dashboard_preferences"`
}

type EmployeeResponse struct {
	ID               int64  `json:"id"`
	UserID           *int64 `json:"user_id"`
	DefaultCompanyID *int64 `json:"default_company_id"`
	FirstName        string `json:"first_name" binding:"omitempty,max=100"`
	LastName         string `json:"last_name" binding:"omitempty,max=100"`
	EmployeeID       string `json:"employee_id" binding:"omitempty,max=50"`
	// IsActive is the employee's status in the session's company. On the
	// cross-company /admin/employees listing it is the account-wide flag.
	IsActive          bool             `json:"is_active"`
	Email             string           `json:"email" binding:"omitempty,max=254"`
	MobilePhone       string           `json:"mobile_phone" binding:"omitempty,max=20"`
	WorkPhone         string           `json:"work_phone" binding:"omitempty,max=20"`
	JobTitle          string           `json:"job_title" binding:"omitempty,max=100"`
	StartDate         *time.Time       `json:"start_date"`
	LeaveDate         *time.Time       `json:"leave_date"`
	BirthDate         *time.Time       `json:"birth_date"`
	HourlyLaborRate   *decimal.Decimal `json:"hourly_labor_rate"`
	IsTechnician      bool             `json:"is_technician"`
	IsVehicleOperator bool             `json:"is_vehicle_operator"`
	// RoleID is the role this employee holds in the session's company, not a
	// property of the person: the same employee can hold a different role, or
	// none, in another company. Null means no role there. Always null on the
	// cross-company /admin/employees listing, which has no session company.
	RoleID *int64 `json:"role_id"`
	// RoleName is that role's name, so a client can show it without reading
	// /roles, which only administrators may.
	RoleName *string `json:"role_name"`
	// IsAccountOwner is read-only and computed from account ownership. It is
	// ignored if sent; ownership is transferred with
	// POST /api/v1/admin/accounts/{id}/set-owner.
	IsAccountOwner       bool            `json:"is_account_owner"`
	LicenseClass         string          `json:"license_class" binding:"omitempty,max=10"`
	LicenseNumber        string          `json:"license_number" binding:"omitempty,max=50"`
	LicenseState         string          `json:"license_state" binding:"omitempty,max=50"`
	LicenseExpiry        *time.Time      `json:"license_expiry"`
	StreetAddress        string          `json:"street_address" binding:"omitempty,max=200"`
	City                 string          `json:"city" binding:"omitempty,max=100"`
	Region               string          `json:"region" binding:"omitempty,max=50"`
	PostalCode           string          `json:"postal_code" binding:"omitempty,max=20"`
	Country              string          `json:"country" binding:"omitempty,max=50"`
	GroupID              *int64          `json:"group_id"`
	CustomFields         json.RawMessage `json:"custom_fields" swaggertype:"object"`
	TablePreferences     json.RawMessage `json:"table_preferences"`
	DashboardPreferences json.RawMessage `json:"dashboard_preferences"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

type EmployeePage struct {
	Data    []EmployeeResponse `json:"data"`
	Total   int64              `json:"total"`
	Limit   int                `json:"limit"`
	Offset  int                `json:"offset"`
	HasNext bool               `json:"has_next"`
}

// SetEmployeePasswordRequest carries a new password for an employee. The
// 8-character floor matches the CLI's setpass rule and Django's serializer.
type SetEmployeePasswordRequest struct {
	Password string `json:"password" binding:"required,min=8,max=128"`
}
