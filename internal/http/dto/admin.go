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

// CompanyOwnerResponse is the employee who currently carries
// is_account_owner among a company's members.
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
