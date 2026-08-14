package dto

// MePermissionsResponse is everything a client needs to render the application
// shell in one request: who the caller is, what their role permits, and which
// companies they can switch to.
type MePermissionsResponse struct {
	Employee       MeEmployee `json:"employee"`
	CompanyID      int64      `json:"company_id"`
	Role           *MeRole    `json:"role"`
	IsAdmin        bool       `json:"is_admin"`
	IsAccountOwner bool       `json:"is_account_owner"`
	// Permissions is the effective answer per module and action, with the
	// empty-object ("whole module") convention and the admin bypass already
	// applied — clients should not have to re-derive either.
	Permissions map[string]map[string]bool `json:"permissions"`
	// Modules lists the modules the caller can read, ready for navigation
	// filtering.
	Modules   []string    `json:"modules"`
	Companies []MeCompany `json:"companies"`
}

// MeEmployee describes the signed-in employee. is_technician and
// is_vehicle_operator are operational flags, not permissions: they say which
// capture flows (assigning labor, logging fuel) apply to this person, which the
// client cannot infer from the permission map.
type MeEmployee struct {
	ID                int64  `json:"id"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Email             string `json:"email"`
	JobTitle          string `json:"job_title"`
	IsActive          bool   `json:"is_active"`
	IsTechnician      bool   `json:"is_technician"`
	IsVehicleOperator bool   `json:"is_vehicle_operator"`
	DefaultCompanyID  *int64 `json:"default_company_id"`
}

type MeRole struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"`
}

type MeCompany struct {
	ID   int64   `json:"id"`
	Name string  `json:"name"`
	Logo *string `json:"logo"`
}
