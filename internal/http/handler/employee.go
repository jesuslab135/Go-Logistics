package handler

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"fleet/internal/db/gen"
	"fleet/internal/http/dto"
	"fleet/internal/http/middleware"
	"fleet/internal/platform/apierr"
	"fleet/internal/platform/paginate"
)

// Status changes to a membership are audited alongside grants and revokes.
const (
	membershipSuspended = "suspended"
	membershipRestored  = "restored"
)

// EmployeeStore scopes employees by company membership (employee_companies m2m),
// creates the membership row transactionally with the employee, and never
// exposes or accepts password_hash (that is owned by the auth flow / CLI).
//
// role_id and is_active on these routes describe the employee's membership in
// the SESSION's company, not the person: the same employee can hold a
// different role, or be suspended, in another company. Delete likewise removes
// them from this company, and deletes the person only when it was their last.
type EmployeeStore struct {
	q    *gen.Queries
	pool *pgxpool.Pool
}

func NewEmployeeStore(q *gen.Queries, pool *pgxpool.Pool) *EmployeeStore {
	return &EmployeeStore{q: q, pool: pool}
}

func (s *EmployeeStore) List(ctx context.Context, p paginate.Params) ([]dto.EmployeeResponse, int64, error) {
	company := middleware.CompanyFromContext(ctx)
	rows, err := s.q.ListEmployees(ctx, gen.ListEmployeesParams{CompanyID: company, Lim: int32(p.Limit), Off: int32(p.Offset)})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountEmployees(ctx, company)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.EmployeeResponse, len(rows))
	for i, r := range rows {
		out[i] = toEmployeeResponse(r)
	}
	if err := applyMemberships(ctx, s.q, company, out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (s *EmployeeStore) Get(ctx context.Context, id int64) (dto.EmployeeResponse, error) {
	company := middleware.CompanyFromContext(ctx)
	r, err := s.q.GetEmployee(ctx, gen.GetEmployeeParams{ID: id, CompanyID: company})
	if err != nil {
		return dto.EmployeeResponse{}, err
	}
	return s.respond(ctx, company, r)
}

func (s *EmployeeStore) Create(ctx context.Context, in dto.CreateEmployeeRequest) (dto.EmployeeResponse, error) {
	// The document has to satisfy whatever this company declared for employees;
	// a company that declared nothing pays one indexed lookup.
	if err := validateCustomFields(ctx, s.q, "employees", in.CustomFields); err != nil {
		return dto.EmployeeResponse{}, err
	}

	company := middleware.CompanyFromContext(ctx)

	// Create grants exactly one membership — the caller's own company — so that
	// is the only default_company_id this request can legitimately set.
	if err := validateDefaultCompany(in.DefaultCompanyID, []int64{company}); err != nil {
		return dto.EmployeeResponse{}, err
	}

	// account_id comes from the caller, never from the body — the same rule
	// company creation follows. This route requires company membership, and
	// platform staff hold none, so a caller who reaches here always has one.
	accountID := middleware.AccountFromContext(ctx)
	if accountID == nil {
		return dto.EmployeeResponse{}, apierr.Forbidden(
			"only a member of a client account can create an employee")
	}

	// A role in the body is only honoured for an administrator of this company.
	// Omitted or null creates the membership with no role, which grants nothing.
	var roleID *int64
	if in.RoleID.Set && in.RoleID.Value != nil {
		if err := authorizeRoleChange(ctx, s.q, 0, company, in.RoleID.Value); err != nil {
			return dto.EmployeeResponse{}, err
		}
		roleID = in.RoleID.Value
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return dto.EmployeeResponse{}, err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)
	now := time.Now().UTC()
	r, err := qtx.CreateEmployee(ctx, createEmployeeParams(in, accountID, now))
	if err != nil {
		return dto.EmployeeResponse{}, err
	}
	if err := qtx.GrantMembership(ctx, gen.GrantMembershipParams{EmployeeID: r.ID, CompanyID: company, RoleID: roleID}); err != nil {
		return dto.EmployeeResponse{}, err
	}
	if !newMembershipActive(in) {
		if _, err := qtx.SetMembershipActive(ctx, gen.SetMembershipActiveParams{IsActive: false, EmployeeID: r.ID, CompanyID: company}); err != nil {
			return dto.EmployeeResponse{}, err
		}
	}
	if roleID != nil {
		if err := recordMembershipChange(ctx, qtx, r.ID, company, membershipGranted, now); err != nil {
			return dto.EmployeeResponse{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return dto.EmployeeResponse{}, err
	}
	return s.respond(ctx, company, r)
}

func (s *EmployeeStore) Update(ctx context.Context, id int64, in dto.UpdateEmployeeRequest) (dto.EmployeeResponse, error) {
	// The document has to satisfy whatever this company declared for employees;
	// a company that declared nothing pays one indexed lookup.
	if err := validateCustomFields(ctx, s.q, "employees", in.CustomFields); err != nil {
		return dto.EmployeeResponse{}, err
	}

	memberships, err := s.q.ListEmployeeCompanyIDs(ctx, id)
	if err != nil {
		return dto.EmployeeResponse{}, err
	}
	if err := validateDefaultCompany(in.DefaultCompanyID, memberships); err != nil {
		return dto.EmployeeResponse{}, err
	}

	company := middleware.CompanyFromContext(ctx)
	now := time.Now().UTC()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return dto.EmployeeResponse{}, err
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	current, err := qtx.GetMembership(ctx, gen.GetMembershipParams{EmployeeID: id, CompanyID: company})
	if err != nil {
		return dto.EmployeeResponse{}, err
	}

	// role_id is only acted on when present in the body and different from the
	// role already held, so a client that echoes the current value back, or
	// never sends the field, needs no administrator rights to save a profile.
	if in.RoleID.Set && !sameRole(current.RoleID, in.RoleID.Value) {
		if err := authorizeRoleChange(ctx, qtx, id, company, in.RoleID.Value); err != nil {
			return dto.EmployeeResponse{}, err
		}
		n, err := qtx.SetMembershipRole(ctx, gen.SetMembershipRoleParams{RoleID: in.RoleID.Value, EmployeeID: id, CompanyID: company})
		if err != nil {
			return dto.EmployeeResponse{}, err
		}
		if n == 0 {
			return dto.EmployeeResponse{}, apierr.NotFound("employee not found")
		}
		if err := recordMembershipChange(ctx, qtx, id, company, membershipGranted, now); err != nil {
			return dto.EmployeeResponse{}, err
		}
	}

	// is_active suspends or restores the membership in this company only.
	if in.IsActive != nil && *in.IsActive != current.IsActive {
		if err := authorizeActiveChange(ctx, qtx, id, *in.IsActive); err != nil {
			return dto.EmployeeResponse{}, err
		}
		if _, err := qtx.SetMembershipActive(ctx, gen.SetMembershipActiveParams{IsActive: *in.IsActive, EmployeeID: id, CompanyID: company}); err != nil {
			return dto.EmployeeResponse{}, err
		}
		action := membershipSuspended
		if *in.IsActive {
			action = membershipRestored
			// A deactivation made while is_active was account-wide (before
			// 000023) would otherwise keep the person locked out everywhere.
			if err := qtx.ActivateEmployee(ctx, gen.ActivateEmployeeParams{ID: id, UpdatedAt: now}); err != nil {
				return dto.EmployeeResponse{}, err
			}
		}
		if err := recordMembershipChange(ctx, qtx, id, company, action, now); err != nil {
			return dto.EmployeeResponse{}, err
		}
	}

	r, err := qtx.UpdateEmployee(ctx, gen.UpdateEmployeeParams{
		ID:                   id,
		CompanyID:            company,
		UserID:               in.UserID,
		DefaultCompanyID:     in.DefaultCompanyID,
		FirstName:            in.FirstName,
		LastName:             in.LastName,
		EmployeeID:           in.EmployeeID,
		Email:                in.Email,
		MobilePhone:          in.MobilePhone,
		WorkPhone:            in.WorkPhone,
		JobTitle:             in.JobTitle,
		StartDate:            in.StartDate,
		LeaveDate:            in.LeaveDate,
		BirthDate:            in.BirthDate,
		HourlyLaborRate:      in.HourlyLaborRate,
		IsTechnician:         in.IsTechnician,
		IsVehicleOperator:    in.IsVehicleOperator,
		LicenseClass:         in.LicenseClass,
		LicenseNumber:        in.LicenseNumber,
		LicenseState:         in.LicenseState,
		LicenseExpiry:        in.LicenseExpiry,
		StreetAddress:        in.StreetAddress,
		City:                 in.City,
		Region:               in.Region,
		PostalCode:           in.PostalCode,
		Country:              in.Country,
		GroupID:              in.GroupID,
		CustomFields:         jsonbOrDefault(in.CustomFields, "{}"),
		TablePreferences:     jsonbOrDefault(in.TablePreferences, "{}"),
		DashboardPreferences: jsonbOrDefault(in.DashboardPreferences, "{}"),
		UpdatedAt:            now,
	})
	if err != nil {
		return dto.EmployeeResponse{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return dto.EmployeeResponse{}, err
	}
	return s.respond(ctx, company, r)
}

// Delete removes the employee from the session's company. Deleting the person
// outright would take them out of every company in the account, including
// ones the caller cannot see, so that only happens when this was their last
// membership — and then fails with a 409 if they have history (labor time,
// issues, fuel entries, inspections) that must keep its author.
func (s *EmployeeStore) Delete(ctx context.Context, id int64) error {
	company := middleware.CompanyFromContext(ctx)

	identity, _ := middleware.IdentityFromContext(ctx)
	if id == identity.EmployeeID {
		return apierr.Conflict("you cannot delete yourself")
	}
	// Deleting the owner would leave the account with none, and with it nobody
	// able to create companies or appoint directors.
	owner, err := isAccountOwner(ctx, s.q, id)
	if err != nil {
		return err
	}
	if owner {
		return apierr.Conflict("the account owner cannot be deleted; transfer ownership first")
	}

	if _, err := s.q.GetMembership(ctx, gen.GetMembershipParams{EmployeeID: id, CompanyID: company}); err != nil {
		return err
	}
	others, err := s.q.CountOtherMemberships(ctx, gen.CountOtherMembershipsParams{EmployeeID: id, CompanyID: company})
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	if err := recordMembershipChange(ctx, qtx, id, company, membershipRevoked, time.Now().UTC()); err != nil {
		return err
	}
	if others > 0 {
		if _, err := qtx.RevokeMembership(ctx, gen.RevokeMembershipParams{EmployeeID: id, CompanyID: company}); err != nil {
			return err
		}
		if err := qtx.ClearDefaultCompanyIf(ctx, gen.ClearDefaultCompanyIfParams{ID: id, CompanyID: company}); err != nil {
			return err
		}
	} else if err := qtx.DeleteEmployee(ctx, gen.DeleteEmployeeParams{ID: id, CompanyID: company}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// respond renders one employee with their membership in company and their
// account ownership filled in.
func (s *EmployeeStore) respond(ctx context.Context, company int64, r gen.Employee) (dto.EmployeeResponse, error) {
	out := []dto.EmployeeResponse{toEmployeeResponse(r)}
	if err := applyMemberships(ctx, s.q, company, out); err != nil {
		return dto.EmployeeResponse{}, err
	}
	return out[0], nil
}

// authorizeRoleChange decides whether the caller may give employeeID the role
// roleID in company. Assigning a role is an administrator's act: without this
// check, anyone allowed to edit employee profiles could make themselves an
// administrator. employeeID is 0 for an employee still being created.
func authorizeRoleChange(ctx context.Context, q *gen.Queries, employeeID, company int64, roleID *int64) error {
	identity, ok := middleware.IdentityFromContext(ctx)
	if !ok || !identity.IsAdmin {
		return apierr.Forbidden("an administrator role is required to assign a role")
	}
	// An administrator removing their own administrator role would lock
	// themselves out of this company. The account owner cannot, because
	// ownership keeps them administrator whatever their role says.
	if employeeID == identity.EmployeeID && !identity.IsAccountOwner {
		return apierr.Validation(map[string]string{"role_id": "you cannot change your own role"})
	}
	if roleID == nil {
		return nil
	}
	belongs, err := q.RoleBelongsToCompany(ctx, gen.RoleBelongsToCompanyParams{RoleID: *roleID, CompanyID: company})
	if err != nil {
		return err
	}
	if !belongs {
		return apierr.Validation(map[string]string{"role_id": "does not belong to this company"})
	}
	return nil
}

// authorizeActiveChange refuses the two deactivations that lock the wrong
// person out: yourself, and the account owner, who would lose the company
// their account owns. Reactivating is never refused.
func authorizeActiveChange(ctx context.Context, q *gen.Queries, employeeID int64, activate bool) error {
	if activate {
		return nil
	}
	identity, _ := middleware.IdentityFromContext(ctx)
	if employeeID == identity.EmployeeID {
		return apierr.Validation(map[string]string{"is_active": "you cannot deactivate yourself"})
	}
	owner, err := isAccountOwner(ctx, q, employeeID)
	if err != nil {
		return err
	}
	if owner {
		return apierr.Validation(map[string]string{"is_active": "the account owner cannot be deactivated"})
	}
	return nil
}

func isAccountOwner(ctx context.Context, q *gen.Queries, employeeID int64) (bool, error) {
	owners, err := q.ListOwnersAmong(ctx, []int64{employeeID})
	if err != nil {
		return false, err
	}
	return slices.Contains(owners, employeeID), nil
}

// recordMembershipChange writes to membership_audit, as every other membership
// write does: a role, a suspension or a removal changes what someone may do.
func recordMembershipChange(ctx context.Context, q *gen.Queries, employeeID, company int64, action string, now time.Time) error {
	actor := middleware.EmployeeFromContext(ctx)
	return q.RecordMembershipChange(ctx, gen.RecordMembershipChangeParams{
		ActorEmployeeID:   &actor,
		SubjectEmployeeID: employeeID,
		CompanyID:         company,
		Action:            action,
		OccurredAt:        now,
	})
}

func sameRole(a, b *int64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// applyMemberships fills in each employee's role, role name and status in
// company, and whether they own their account, with one read for each rather
// than one per employee.
func applyMemberships(ctx context.Context, q *gen.Queries, company int64, out []dto.EmployeeResponse) error {
	if len(out) == 0 {
		return nil
	}
	ids := make([]int64, len(out))
	for i := range out {
		ids[i] = out[i].ID
	}
	rows, err := q.ListCompanyMemberships(ctx, gen.ListCompanyMembershipsParams{CompanyID: company, EmployeeIds: ids})
	if err != nil {
		return err
	}
	byEmployee := make(map[int64]gen.ListCompanyMembershipsRow, len(rows))
	for _, r := range rows {
		byEmployee[r.EmployeeID] = r
	}
	for i := range out {
		if m, ok := byEmployee[out[i].ID]; ok {
			out[i].RoleID = m.RoleID
			out[i].RoleName = m.RoleName
			out[i].IsActive = out[i].IsActive && m.IsActive
		}
	}
	return applyAccountOwners(ctx, q, out)
}

// applyAccountOwners marks the employees who own their account. Ownership
// lives on account.owner_employee_id, not on the employee row.
func applyAccountOwners(ctx context.Context, q *gen.Queries, out []dto.EmployeeResponse) error {
	if len(out) == 0 {
		return nil
	}
	ids := make([]int64, len(out))
	for i := range out {
		ids[i] = out[i].ID
	}
	owners, err := q.ListOwnersAmong(ctx, ids)
	if err != nil {
		return err
	}
	for i := range out {
		out[i].IsAccountOwner = slices.Contains(owners, out[i].ID)
	}
	return nil
}

// validateDefaultCompany rejects a default_company_id the employee has no
// employee_companies row for. Login scopes a session to this value, so an
// unvalidated one is a cross-tenant token waiting to be issued.
func validateDefaultCompany(defaultCompanyID *int64, memberships []int64) error {
	if defaultCompanyID == nil {
		return nil
	}
	if *defaultCompanyID > 0 && slices.Contains(memberships, *defaultCompanyID) {
		return nil
	}
	return apierr.Validation(map[string]string{
		"default_company_id": "must be a company this employee belongs to",
	})
}

// newMembershipActive is the status the new employee's membership starts with.
// is_active is a pointer so an explicit false wins over the active default.
func newMembershipActive(in dto.CreateEmployeeRequest) bool {
	return boolOrDefault(in.IsActive, true)
}

// createEmployeeParams creates the employee row always active: deactivation
// belongs to the membership (see newMembershipActive), and the row's own
// is_active is only the old account-wide switch.
func createEmployeeParams(in dto.CreateEmployeeRequest, accountID *int64, now time.Time) gen.CreateEmployeeParams {
	return gen.CreateEmployeeParams{
		AccountID:            accountID,
		UserID:               in.UserID,
		DefaultCompanyID:     in.DefaultCompanyID,
		FirstName:            in.FirstName,
		LastName:             in.LastName,
		EmployeeID:           in.EmployeeID,
		IsActive:             true,
		Email:                in.Email,
		MobilePhone:          in.MobilePhone,
		WorkPhone:            in.WorkPhone,
		JobTitle:             in.JobTitle,
		StartDate:            in.StartDate,
		LeaveDate:            in.LeaveDate,
		BirthDate:            in.BirthDate,
		HourlyLaborRate:      in.HourlyLaborRate,
		IsTechnician:         in.IsTechnician,
		IsVehicleOperator:    in.IsVehicleOperator,
		LicenseClass:         in.LicenseClass,
		LicenseNumber:        in.LicenseNumber,
		LicenseState:         in.LicenseState,
		LicenseExpiry:        in.LicenseExpiry,
		StreetAddress:        in.StreetAddress,
		City:                 in.City,
		Region:               in.Region,
		PostalCode:           in.PostalCode,
		Country:              in.Country,
		GroupID:              in.GroupID,
		CustomFields:         jsonbOrDefault(in.CustomFields, "{}"),
		TablePreferences:     jsonbOrDefault(in.TablePreferences, "{}"),
		DashboardPreferences: jsonbOrDefault(in.DashboardPreferences, "{}"),
		UpdatedAt:            now,
	}
}

// toEmployeeResponse renders the employee row alone. RoleID, RoleName, the
// per-company IsActive and IsAccountOwner are not columns of it;
// applyMemberships and applyAccountOwners fill them.
func toEmployeeResponse(r gen.Employee) dto.EmployeeResponse {
	return dto.EmployeeResponse{
		ID:                   r.ID,
		UserID:               r.UserID,
		DefaultCompanyID:     r.DefaultCompanyID,
		FirstName:            r.FirstName,
		LastName:             r.LastName,
		EmployeeID:           r.EmployeeID,
		IsActive:             r.IsActive,
		Email:                r.Email,
		MobilePhone:          r.MobilePhone,
		WorkPhone:            r.WorkPhone,
		JobTitle:             r.JobTitle,
		StartDate:            r.StartDate,
		LeaveDate:            r.LeaveDate,
		BirthDate:            r.BirthDate,
		HourlyLaborRate:      r.HourlyLaborRate,
		IsTechnician:         r.IsTechnician,
		IsVehicleOperator:    r.IsVehicleOperator,
		LicenseClass:         r.LicenseClass,
		LicenseNumber:        r.LicenseNumber,
		LicenseState:         r.LicenseState,
		LicenseExpiry:        r.LicenseExpiry,
		StreetAddress:        r.StreetAddress,
		City:                 r.City,
		Region:               r.Region,
		PostalCode:           r.PostalCode,
		Country:              r.Country,
		GroupID:              r.GroupID,
		CustomFields:         json.RawMessage(r.CustomFields),
		TablePreferences:     json.RawMessage(r.TablePreferences),
		DashboardPreferences: json.RawMessage(r.DashboardPreferences),
		UpdatedAt:            r.UpdatedAt,
	}
}
