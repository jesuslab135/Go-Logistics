package dto

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

type CreateEmployeeRequest struct {
	UserID               *int64           `json:"user_id"`
	DefaultCompanyID     *int64           `json:"default_company_id"`
	FirstName            string           `json:"first_name" binding:"omitempty,max=100"`
	LastName             string           `json:"last_name" binding:"omitempty,max=100"`
	EmployeeID           string           `json:"employee_id" binding:"omitempty,max=50"`
	RoleID               *int64           `json:"role_id"`
	IsActive             bool             `json:"is_active"`
	Email                string           `json:"email" binding:"omitempty,max=254"`
	MobilePhone          string           `json:"mobile_phone" binding:"omitempty,max=20"`
	WorkPhone            string           `json:"work_phone" binding:"omitempty,max=20"`
	JobTitle             string           `json:"job_title" binding:"omitempty,max=100"`
	StartDate            *time.Time       `json:"start_date"`
	LeaveDate            *time.Time       `json:"leave_date"`
	BirthDate            *time.Time       `json:"birth_date"`
	HourlyLaborRate      *decimal.Decimal `json:"hourly_labor_rate"`
	IsTechnician         bool             `json:"is_technician"`
	IsVehicleOperator    bool             `json:"is_vehicle_operator"`
	IsAccountOwner       bool             `json:"is_account_owner"`
	LicenseClass         string           `json:"license_class" binding:"omitempty,max=10"`
	LicenseNumber        string           `json:"license_number" binding:"omitempty,max=50"`
	LicenseState         string           `json:"license_state" binding:"omitempty,max=50"`
	LicenseExpiry        *time.Time       `json:"license_expiry"`
	StreetAddress        string           `json:"street_address" binding:"omitempty,max=200"`
	City                 string           `json:"city" binding:"omitempty,max=100"`
	Region               string           `json:"region" binding:"omitempty,max=50"`
	PostalCode           string           `json:"postal_code" binding:"omitempty,max=20"`
	Country              string           `json:"country" binding:"omitempty,max=50"`
	GroupID              *int64           `json:"group_id"`
	CustomFields         json.RawMessage  `json:"custom_fields"`
	TablePreferences     json.RawMessage  `json:"table_preferences"`
	DashboardPreferences json.RawMessage  `json:"dashboard_preferences"`
}

type UpdateEmployeeRequest struct {
	UserID               *int64           `json:"user_id"`
	DefaultCompanyID     *int64           `json:"default_company_id"`
	FirstName            string           `json:"first_name" binding:"omitempty,max=100"`
	LastName             string           `json:"last_name" binding:"omitempty,max=100"`
	EmployeeID           string           `json:"employee_id" binding:"omitempty,max=50"`
	RoleID               *int64           `json:"role_id"`
	IsActive             bool             `json:"is_active"`
	Email                string           `json:"email" binding:"omitempty,max=254"`
	MobilePhone          string           `json:"mobile_phone" binding:"omitempty,max=20"`
	WorkPhone            string           `json:"work_phone" binding:"omitempty,max=20"`
	JobTitle             string           `json:"job_title" binding:"omitempty,max=100"`
	StartDate            *time.Time       `json:"start_date"`
	LeaveDate            *time.Time       `json:"leave_date"`
	BirthDate            *time.Time       `json:"birth_date"`
	HourlyLaborRate      *decimal.Decimal `json:"hourly_labor_rate"`
	IsTechnician         bool             `json:"is_technician"`
	IsVehicleOperator    bool             `json:"is_vehicle_operator"`
	IsAccountOwner       bool             `json:"is_account_owner"`
	LicenseClass         string           `json:"license_class" binding:"omitempty,max=10"`
	LicenseNumber        string           `json:"license_number" binding:"omitempty,max=50"`
	LicenseState         string           `json:"license_state" binding:"omitempty,max=50"`
	LicenseExpiry        *time.Time       `json:"license_expiry"`
	StreetAddress        string           `json:"street_address" binding:"omitempty,max=200"`
	City                 string           `json:"city" binding:"omitempty,max=100"`
	Region               string           `json:"region" binding:"omitempty,max=50"`
	PostalCode           string           `json:"postal_code" binding:"omitempty,max=20"`
	Country              string           `json:"country" binding:"omitempty,max=50"`
	GroupID              *int64           `json:"group_id"`
	CustomFields         json.RawMessage  `json:"custom_fields"`
	TablePreferences     json.RawMessage  `json:"table_preferences"`
	DashboardPreferences json.RawMessage  `json:"dashboard_preferences"`
}

type EmployeeResponse struct {
	ID                   int64            `json:"id"`
	UserID               *int64           `json:"user_id"`
	DefaultCompanyID     *int64           `json:"default_company_id"`
	FirstName            string           `json:"first_name" binding:"omitempty,max=100"`
	LastName             string           `json:"last_name" binding:"omitempty,max=100"`
	EmployeeID           string           `json:"employee_id" binding:"omitempty,max=50"`
	RoleID               *int64           `json:"role_id"`
	IsActive             bool             `json:"is_active"`
	Email                string           `json:"email" binding:"omitempty,max=254"`
	MobilePhone          string           `json:"mobile_phone" binding:"omitempty,max=20"`
	WorkPhone            string           `json:"work_phone" binding:"omitempty,max=20"`
	JobTitle             string           `json:"job_title" binding:"omitempty,max=100"`
	StartDate            *time.Time       `json:"start_date"`
	LeaveDate            *time.Time       `json:"leave_date"`
	BirthDate            *time.Time       `json:"birth_date"`
	HourlyLaborRate      *decimal.Decimal `json:"hourly_labor_rate"`
	IsTechnician         bool             `json:"is_technician"`
	IsVehicleOperator    bool             `json:"is_vehicle_operator"`
	IsAccountOwner       bool             `json:"is_account_owner"`
	LicenseClass         string           `json:"license_class" binding:"omitempty,max=10"`
	LicenseNumber        string           `json:"license_number" binding:"omitempty,max=50"`
	LicenseState         string           `json:"license_state" binding:"omitempty,max=50"`
	LicenseExpiry        *time.Time       `json:"license_expiry"`
	StreetAddress        string           `json:"street_address" binding:"omitempty,max=200"`
	City                 string           `json:"city" binding:"omitempty,max=100"`
	Region               string           `json:"region" binding:"omitempty,max=50"`
	PostalCode           string           `json:"postal_code" binding:"omitempty,max=20"`
	Country              string           `json:"country" binding:"omitempty,max=50"`
	GroupID              *int64           `json:"group_id"`
	CustomFields         json.RawMessage  `json:"custom_fields"`
	TablePreferences     json.RawMessage  `json:"table_preferences"`
	DashboardPreferences json.RawMessage  `json:"dashboard_preferences"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

type EmployeePage struct {
	Data    []EmployeeResponse `json:"data"`
	Total   int64              `json:"total"`
	Limit   int                `json:"limit"`
	Offset  int                `json:"offset"`
	HasNext bool               `json:"has_next"`
}
