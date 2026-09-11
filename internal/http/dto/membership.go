package dto

// CompanyRoleGrant gives an employee access to one company with one role.
// RoleID is optional: a membership with no role grants nothing, which is a
// legitimate state — associated with the company, permitted nothing in it.
type CompanyRoleGrant struct {
	CompanyID int64  `json:"company_id"`
	RoleID    *int64 `json:"role_id"`
}

// ReplaceAccountEmployeeCompaniesRequest replaces the whole set. An empty list
// revokes every membership, which is how access is removed.
type ReplaceAccountEmployeeCompaniesRequest struct {
	Grants []CompanyRoleGrant `json:"grants"`
}

type AccountEmployeeMembership struct {
	CompanyID   int64  `json:"company_id"`
	CompanyName string `json:"company_name"`
	RoleID      *int64 `json:"role_id"`
	RoleName    string `json:"role_name,omitempty"`
	RoleIsAdmin bool   `json:"role_is_admin"`
	// IsActive is false when the person is suspended in this company; they
	// keep their other memberships.
	IsActive bool `json:"is_active"`
}

type AccountEmployeeResponse struct {
	EmployeeID  int64                       `json:"employee_id"`
	FirstName   string                      `json:"first_name"`
	LastName    string                      `json:"last_name"`
	Email       string                      `json:"email"`
	IsActive    bool                        `json:"is_active"`
	Memberships []AccountEmployeeMembership `json:"memberships"`
}
