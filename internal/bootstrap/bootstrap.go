// Package bootstrap seeds the rows a company needs before the product is usable.
//
// Django did this in two places: CompanyViewSet.perform_create created the admin
// role, the default work order statuses and the owner's membership inside one
// transaction, while asset statuses were seeded separately by the
// seed_asset_statuses management command. Both paths live here so the HTTP
// handler and the CLI share one definition.
package bootstrap

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"fleet/internal/db/gen"
)

// AdminRoleName is the role Django created for every new company.
const AdminRoleName = "Administrador"

// WarehouseRoleName is the second canonical role, ported from
// api/management/commands/seed_warehouse_role.py. Django created it out-of-band
// via a management command, which meant a fresh company had no non-admin role at
// all until someone remembered to run it; seeding it here removes that step.
const WarehouseRoleName = "Almacén"

// DefaultRoles are seeded for every company. An empty permissions object grants
// the whole module (see middleware.DecodePermissions), which is why the admin
// role carries `{}` — its access comes from is_admin, not from an enumeration.
var DefaultRoles = []struct {
	Name        string
	IsAdmin     bool
	Permissions string
}{
	{AdminRoleName, true, `{}`},
	// WAREHOUSE_PERMISSIONS from seed_warehouse_role.py, verbatim. "approve"
	// is not one of the four method-derived actions; it gated the tire
	// assignment approve/reject endpoints.
	{WarehouseRoleName, false, `{
        "tire_approvals": {"read": true, "create": true, "update": true, "delete": true, "approve": true},
        "tires":          {"read": true},
        "inventory":      {"read": true, "update": true},
        "parts":          {"read": true},
        "assets":         {"read": true}
    }`},
}

// DefaultAssetStatuses ports DEFAULT_ASSET_STATUSES from
// api/management/commands/seed_asset_statuses.py.
var DefaultAssetStatuses = []struct {
	Name      string
	ColorCode string
}{
	{"Operativo", "#28a745"},
	{"Taller", "#ffc107"},
	{"Safety", "#fd7e14"},
	{"Compliance", "#6f42c1"},
	{"Sin Placas", "#dc3545"},
}

// DefaultWorkOrderStatuses ports DEFAULT_WORK_ORDER_STATUSES from
// api/Viewsets/company_viewset.py.
var DefaultWorkOrderStatuses = []struct {
	Name             string
	Color            string
	Position         int32
	IsDefault        bool
	MarksAsCompleted bool
}{
	{"Abierta", "#3B82F6", 0, true, false},
	{"En Fila", "#64748B", 1, false, false},
	{"Diagnóstico", "#0EA5E9", 2, false, false},
	{"En Proceso", "#F59E0B", 3, false, false},
	{"Pendiente Refacciones", "#A855F7", 4, false, false},
	{"Esperando Partes", "#C084FC", 5, false, false},
	{"Liberado", "#22C55E", 6, false, false},
	{"Completada", "#10B981", 7, false, true},
}

// Company seeds companyID and, when employeeID is non-zero, links that employee
// to it as Django's perform_create did: membership row, plus role and default
// company where the employee has none. It is idempotent, so it can be re-run
// against a company created before this existed.
func Company(ctx context.Context, q *gen.Queries, companyID, employeeID int64) error {
	var adminRole gen.Role
	for _, r := range DefaultRoles {
		role, err := ensureRole(ctx, q, companyID, r.Name, r.IsAdmin, r.Permissions)
		if err != nil {
			return err
		}
		if r.Name == AdminRoleName {
			adminRole = role
		}
	}

	for _, st := range DefaultWorkOrderStatuses {
		if err := q.SeedWorkOrderStatus(ctx, gen.SeedWorkOrderStatusParams{
			CompanyID:        companyID,
			Name:             st.Name,
			Color:            st.Color,
			IsDefault:        st.IsDefault,
			MarksAsCompleted: st.MarksAsCompleted,
			Position:         st.Position,
		}); err != nil {
			return err
		}
	}

	for _, st := range DefaultAssetStatuses {
		if err := q.SeedAssetStatus(ctx, gen.SeedAssetStatusParams{
			CompanyID: companyID,
			Name:      st.Name,
			ColorCode: st.ColorCode,
		}); err != nil {
			return err
		}
	}

	if employeeID == 0 {
		return nil
	}
	if err := q.AddEmployeeCompany(ctx, gen.AddEmployeeCompanyParams{
		EmployeeID: employeeID,
		CompanyID:  companyID,
	}); err != nil {
		return err
	}
	return q.BootstrapEmployeeCompany(ctx, gen.BootstrapEmployeeCompanyParams{
		ID:        employeeID,
		RoleID:    &adminRole.ID,
		CompanyID: &companyID,
		UpdatedAt: time.Now().UTC(),
	})
}

// ensureRole is idempotent by name, so re-running the bootstrap never disturbs a
// role an operator has since edited.
func ensureRole(ctx context.Context, q *gen.Queries, companyID int64, name string, isAdmin bool, permissions string) (gen.Role, error) {
	role, err := q.FindRoleByName(ctx, gen.FindRoleByNameParams{CompanyID: companyID, Name: name})
	if err == nil {
		return role, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return gen.Role{}, err
	}
	return q.CreateRole(ctx, gen.CreateRoleParams{
		CompanyID:   companyID,
		Name:        name,
		IsAdmin:     isAdmin,
		Permissions: []byte(permissions),
	})
}
