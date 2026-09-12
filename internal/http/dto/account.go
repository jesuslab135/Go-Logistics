package dto

import "time"

// CreateAccountRequest provisions a client and its owner in one act. The
// password is accepted here rather than in a follow-up call because a
// two-step flow can leave an account whose owner exists but cannot log in.
type CreateAccountRequest struct {
	Name           string `json:"name"`
	OwnerFirstName string `json:"owner_first_name"`
	OwnerLastName  string `json:"owner_last_name"`
	OwnerEmail     string `json:"owner_email"`
	// OwnerPassword is never echoed back and never logged.
	OwnerPassword string `json:"owner_password"`
}

type AccountResponse struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	OwnerEmployeeID *int64    `json:"owner_employee_id"`
	OwnerEmail      string    `json:"owner_email,omitempty"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	CompanyCount    int64     `json:"company_count"`
	EmployeeCount   int64     `json:"employee_count"`
}

type AccountPage struct {
	Data    []AccountResponse `json:"data"`
	Total   int64             `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	HasNext bool              `json:"has_next"`
}

// AccountCompanyResponse is one company of the caller's account.
type AccountCompanyResponse struct {
	ID   int64   `json:"id"`
	Name string  `json:"name"`
	Logo *string `json:"logo"`
}

type SetAccountOwnerRequest struct {
	EmployeeID int64 `json:"employee_id" binding:"required,min=1"`
}
