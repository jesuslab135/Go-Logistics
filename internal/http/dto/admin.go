package dto

// ReplaceEmployeeCompaniesRequest is a full overwrite of an employee's
// employee_companies rows. company_ids must be non-empty: an employee with no
// membership cannot log in at all, so deactivation belongs on is_active, not
// here. default_company_id must be null or one of company_ids.
type ReplaceEmployeeCompaniesRequest struct {
	CompanyIDs       []int64 `json:"company_ids" binding:"required,dive,min=1"`
	DefaultCompanyID *int64  `json:"default_company_id"`
}

// EmployeeCompaniesResponse is the employee's membership set after the write,
// normalized (sorted, de-duplicated).
type EmployeeCompaniesResponse struct {
	EmployeeID       int64   `json:"employee_id"`
	CompanyIDs       []int64 `json:"company_ids"`
	DefaultCompanyID *int64  `json:"default_company_id"`
}

// CompanyOwnerResponse is the owner of the account a company belongs to.
type CompanyOwnerResponse struct {
	EmployeeID int64  `json:"employee_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	JobTitle   string `json:"job_title"`
}

// CompanyOwnerEnvelope wraps a nullable owner so "this company has no owner"
// is a 200 with owner:null rather than a 404 the frontend has to special-case.
type CompanyOwnerEnvelope struct {
	Owner *CompanyOwnerResponse `json:"owner"`
}

// SetCompanyOwnerRequest names the employee to make the company's owner.
type SetCompanyOwnerRequest struct {
	EmployeeID int64 `json:"employee_id" binding:"required,min=1"`
}

// AdminCompanyResponse is a company as the platform sees it: the tenant fields
// plus the client it belongs to, which the tenant-facing CompanyResponse leaves
// out because a tenant only ever sees its own. Embedded, so the JSON is flat.
type AdminCompanyResponse struct {
	CompanyResponse
	AccountID     int64  `json:"account_id"`
	AccountName   string `json:"account_name"`
	EmployeeCount int64  `json:"employee_count"`
}

type AdminCompanyPage struct {
	Data    []AdminCompanyResponse `json:"data"`
	Total   int64                  `json:"total"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
	HasNext bool                   `json:"has_next"`
}

// AdminEmployeeResponse is an employee as the platform sees it. account_id is
// null for platform staff, who belong to no client; the tenant-facing
// EmployeeResponse has neither field because a tenant only sees its own people.
type AdminEmployeeResponse struct {
	EmployeeResponse
	AccountID       *int64 `json:"account_id"`
	IsPlatformAdmin bool   `json:"is_platform_admin"`
}

type AdminEmployeePage struct {
	Data    []AdminEmployeeResponse `json:"data"`
	Total   int64                   `json:"total"`
	Limit   int                     `json:"limit"`
	Offset  int                     `json:"offset"`
	HasNext bool                    `json:"has_next"`
}
