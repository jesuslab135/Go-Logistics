package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"fleet/internal/db/gen"
)

// defaultRole describes a role to seed when a new company is created.
type defaultRole struct {
	Name        string
	IsAdmin     bool
	Permissions map[string]any
}

// DefaultRoles returns the 6 roles seeded on every new company. The payloads
// match docs/FRONTEND-HANDOFF/03-ROLE-TEMPLATES.md exactly.
func DefaultRoles() []defaultRole {
	return []defaultRole{
		{
			Name:    "Super Admin",
			IsAdmin: true,
			Permissions: map[string]any{
				"navigation": []string{
					"dashboard",
					"assets", "assets.all", "assets.makes", "assets.models", "assets.types_statuses",
					"maintenance", "maintenance.work_orders", "maintenance.service_entries",
					"maintenance.service_reminders", "maintenance.locations",
					"tires", "tires.inventory", "tires.models", "tires.axle_templates",
					"tires.installations", "tires.assignment_requests",
					"fuel", "fuel.entries", "fuel.types",
					"inspections", "inspections.forms", "inspections.submissions",
					"issues", "issues.board", "issues.priorities", "issues.faults",
					"parts_inventory", "parts_inventory.catalog", "parts_inventory.warehouses",
					"parts_inventory.journal", "parts_inventory.reasons",
					"procurement", "procurement.purchase_orders", "procurement.vendors",
					"organization", "organization.employees", "organization.roles", "organization.groups",
					"settings", "settings.company", "settings.wo_statuses", "settings.fuel_types",
					"settings.units", "settings.warranty", "settings.mileage_goals",
				},
				"models": []string{
					"asset", "vehicle", "trailer", "asset_type", "asset_status", "catalog_option",
					"asset_trailer_assignment", "vehicle_make", "vehicle_model",
					"work_order", "work_order_status", "work_order_line_item", "work_order_sub_line_item",
					"labor_time_entry", "work_order_status_log", "location",
					"service_entry", "service_entry_line_item", "service_task", "service_task_part", "service_reminder",
					"tire", "tire_model", "axle_template", "axle_definition", "wheel_position_definition",
					"tire_installation", "tire_inspection", "tire_mount_log", "tire_assignment_request", "vehicle_axle_config",
					"fuel_entry", "fuel_type", "fuel_comment", "fuel_photo",
					"inspection_form", "inspection_form_item", "inspection_submission", "inspection_submission_item",
					"issue", "issue_priority", "fault",
					"part", "part_category", "part_manufacturer", "measurement_unit",
					"part_location", "part_location_detail", "inventory_journal_entry", "inventory_adjustment_reason",
					"purchase_order", "purchase_order_line_item", "vendor",
					"company", "employee", "role", "group",
					"comment", "media", "warranty", "weekly_mileage_goal",
				},
				"actions": map[string]any{},
			},
		},
		{
			Name:    "Fleet Manager",
			IsAdmin: false,
			Permissions: map[string]any{
				"navigation": []string{
					"dashboard",
					"assets", "assets.all", "assets.makes", "assets.models", "assets.types_statuses",
					"maintenance", "maintenance.work_orders", "maintenance.service_entries",
					"maintenance.service_reminders", "maintenance.locations",
					"tires", "tires.inventory", "tires.models", "tires.axle_templates",
					"tires.installations", "tires.assignment_requests",
					"fuel", "fuel.entries", "fuel.types",
					"inspections", "inspections.forms", "inspections.submissions",
					"issues", "issues.board", "issues.priorities", "issues.faults",
					"parts_inventory", "parts_inventory.catalog", "parts_inventory.warehouses",
					"parts_inventory.journal", "parts_inventory.reasons",
					"procurement", "procurement.purchase_orders", "procurement.vendors",
					"organization", "organization.employees", "organization.roles", "organization.groups",
				},
				"models": []string{
					"asset", "vehicle", "trailer", "asset_type", "asset_status", "catalog_option",
					"asset_trailer_assignment", "vehicle_make", "vehicle_model",
					"work_order", "work_order_status", "work_order_line_item", "work_order_sub_line_item",
					"labor_time_entry", "work_order_status_log", "location",
					"service_entry", "service_entry_line_item", "service_task", "service_task_part", "service_reminder",
					"tire", "tire_model", "axle_template", "axle_definition", "wheel_position_definition",
					"tire_installation", "tire_inspection", "tire_mount_log", "tire_assignment_request", "vehicle_axle_config",
					"fuel_entry", "fuel_type", "fuel_comment", "fuel_photo",
					"inspection_form", "inspection_form_item", "inspection_submission", "inspection_submission_item",
					"issue", "issue_priority", "fault",
					"part", "part_category", "part_manufacturer", "measurement_unit",
					"part_location", "part_location_detail", "inventory_journal_entry", "inventory_adjustment_reason",
					"purchase_order", "purchase_order_line_item", "vendor",
					"employee", "role", "group",
					"comment", "media", "warranty", "weekly_mileage_goal",
				},
				"actions": map[string]any{},
			},
		},
		{
			Name:    "Dispatcher",
			IsAdmin: false,
			Permissions: map[string]any{
				"navigation": []string{
					"dashboard",
					"assets", "assets.all",
					"maintenance", "maintenance.work_orders", "maintenance.service_entries",
					"maintenance.service_reminders", "maintenance.locations",
				},
				"models": []string{
					"asset", "vehicle", "trailer",
					"work_order", "work_order_status", "work_order_line_item", "work_order_sub_line_item",
					"labor_time_entry", "work_order_status_log", "location",
					"service_entry", "service_entry_line_item", "service_task", "service_reminder",
				},
				"actions": map[string]any{},
			},
		},
		{
			Name:    "Technician",
			IsAdmin: false,
			Permissions: map[string]any{
				"navigation": []string{
					"dashboard",
					"maintenance", "maintenance.work_orders", "maintenance.service_entries",
					"maintenance.service_reminders",
					"tires", "tires.inventory", "tires.installations",
					"inspections", "inspections.submissions",
					"issues", "issues.board",
				},
				"models": []string{
					"work_order", "work_order_line_item", "work_order_sub_line_item",
					"labor_time_entry", "work_order_status_log",
					"service_entry", "service_entry_line_item", "service_task", "service_reminder",
					"tire", "tire_installation", "tire_inspection",
					"inspection_submission", "inspection_submission_item",
					"issue", "issue_priority", "fault",
				},
				"actions": map[string]any{},
			},
		},
		{
			Name:    "Parts Manager",
			IsAdmin: false,
			Permissions: map[string]any{
				"navigation": []string{
					"dashboard",
					"maintenance", "maintenance.work_orders",
					"parts_inventory", "parts_inventory.catalog", "parts_inventory.warehouses",
					"parts_inventory.journal", "parts_inventory.reasons",
					"procurement", "procurement.purchase_orders", "procurement.vendors",
				},
				"models": []string{
					"work_order",
					"part", "part_category", "part_manufacturer", "measurement_unit",
					"part_location", "part_location_detail", "inventory_journal_entry", "inventory_adjustment_reason",
					"purchase_order", "purchase_order_line_item", "vendor",
				},
				"actions": map[string]any{},
			},
		},
		{
			Name:    "Driver",
			IsAdmin: false,
			Permissions: map[string]any{
				"navigation": []string{
					"dashboard",
					"fuel", "fuel.entries",
					"inspections", "inspections.submissions",
					"issues", "issues.board",
				},
				"models": []string{
					"fuel_entry",
					"inspection_submission", "inspection_submission_item",
					"issue",
				},
				"actions": map[string]any{},
			},
		},
	}
}

// SeedDefaultRoles inserts the 6 default roles for a company. Called within a
// transaction during company creation (HTTP API or CLI).
func SeedDefaultRoles(ctx context.Context, q *gen.Queries, companyID int64) error {
	for _, r := range DefaultRoles() {
		b, err := json.Marshal(r.Permissions)
		if err != nil {
			return fmt.Errorf("marshaling permissions for role %q: %w", r.Name, err)
		}
		if _, err := q.CreateRole(ctx, gen.CreateRoleParams{
			CompanyID:   companyID,
			Name:        r.Name,
			IsAdmin:     r.IsAdmin,
			Permissions: b,
		}); err != nil {
			return fmt.Errorf("creating role %q: %w", r.Name, err)
		}
	}
	return nil
}
