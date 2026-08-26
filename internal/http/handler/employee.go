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

// EmployeeStore scopes employees by company membership (employee_companies m2m),
// creates the membership row transactionally with the employee, and never
// exposes or accepts password_hash (that is owned by the auth flow / CLI).
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
	return out, total, nil
}

func (s *EmployeeStore) Get(ctx context.Context, id int64) (dto.EmployeeResponse, error) {
	r, err := s.q.GetEmployee(ctx, gen.GetEmployeeParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
	if err != nil {
		return dto.EmployeeResponse{}, err
	}
	return toEmployeeResponse(r), nil
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

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return dto.EmployeeResponse{}, err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)
	r, err := qtx.CreateEmployee(ctx, createEmployeeParams(in, time.Now().UTC()))
	if err != nil {
		return dto.EmployeeResponse{}, err
	}
	if err := qtx.AddEmployeeCompany(ctx, gen.AddEmployeeCompanyParams{EmployeeID: r.ID, CompanyID: company}); err != nil {
		return dto.EmployeeResponse{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return dto.EmployeeResponse{}, err
	}
	return toEmployeeResponse(r), nil
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

	r, err := s.q.UpdateEmployee(ctx, gen.UpdateEmployeeParams{
		ID:                   id,
		CompanyID:            middleware.CompanyFromContext(ctx),
		UserID:               in.UserID,
		DefaultCompanyID:     in.DefaultCompanyID,
		FirstName:            in.FirstName,
		LastName:             in.LastName,
		EmployeeID:           in.EmployeeID,
		RoleID:               in.RoleID,
		IsActive:             in.IsActive,
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
		IsAccountOwner:       in.IsAccountOwner,
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
		UpdatedAt:            time.Now().UTC(),
	})
	if err != nil {
		return dto.EmployeeResponse{}, err
	}
	return toEmployeeResponse(r), nil
}

func (s *EmployeeStore) Delete(ctx context.Context, id int64) error {
	return s.q.DeleteEmployee(ctx, gen.DeleteEmployeeParams{ID: id, CompanyID: middleware.CompanyFromContext(ctx)})
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

func createEmployeeParams(in dto.CreateEmployeeRequest, now time.Time) gen.CreateEmployeeParams {
	return gen.CreateEmployeeParams{
		UserID:               in.UserID,
		DefaultCompanyID:     in.DefaultCompanyID,
		FirstName:            in.FirstName,
		LastName:             in.LastName,
		EmployeeID:           in.EmployeeID,
		RoleID:               in.RoleID,
		IsActive:             boolOrDefault(in.IsActive, true),
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
		IsAccountOwner:       in.IsAccountOwner,
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

func toEmployeeResponse(r gen.Employee) dto.EmployeeResponse {
	return dto.EmployeeResponse{
		ID:                   r.ID,
		UserID:               r.UserID,
		DefaultCompanyID:     r.DefaultCompanyID,
		FirstName:            r.FirstName,
		LastName:             r.LastName,
		EmployeeID:           r.EmployeeID,
		RoleID:               r.RoleID,
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
		IsAccountOwner:       r.IsAccountOwner,
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
