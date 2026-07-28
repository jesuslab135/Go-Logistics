package dto

import "encoding/json"

// MePermissionsResponse is the payload returned by GET /api/v1/me/permissions.
// It merges the employee's role permissions with their employee-level flags so
// the frontend can build the sidebar and set route guards in a single request.
type MePermissionsResponse struct {
	UserID     *int64                `json:"user_id"`
	EmployeeID int64                 `json:"employee_id"`
	CompanyID  int64                 `json:"company_id"`
	Role       MePermissionsRole     `json:"role"`
	Employee   MePermissionsEmployee `json:"employee"`
}

type MePermissionsRole struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	IsAdmin     bool            `json:"is_admin"`
	Permissions json.RawMessage `json:"permissions"`
}

type MePermissionsEmployee struct {
	IsTechnician      bool `json:"is_technician"`
	IsVehicleOperator bool `json:"is_vehicle_operator"`
	IsAccountOwner    bool `json:"is_account_owner"`
}
