package middleware

import (
	"net/http"
	"slices"
	"testing"
)

// /me/permissions reports on the modules in this list, so a module missing
// from it is one the frontend is never told about.
func TestReportsIsARegisteredModule(t *testing.T) {
	if !slices.Contains(Modules, "reports") {
		t.Fatal(`Modules does not list "reports"`)
	}
	if !slices.IsSorted(Modules) {
		t.Fatalf("Modules is kept in alphabetical order: %v", Modules)
	}
}

func TestReportsPermissions(t *testing.T) {
	admin := Identity{IsActive: true, IsAdmin: true}
	if !admin.Can("reports", http.MethodGet) || !admin.Can("reports", http.MethodDelete) {
		t.Fatal("an administrator must reach reports without a grant")
	}

	reader := Identity{IsActive: true, HasRole: true, Permissions: map[string]ModulePermissions{
		"reports": {Actions: map[string]bool{"read": true}},
	}}
	if !reader.Can("reports", http.MethodGet) {
		t.Fatal("reports/read must allow a download")
	}
	if reader.Can("reports", http.MethodPost) || reader.Can("reports", http.MethodDelete) {
		t.Fatal("reports/read must not allow changing the recipients")
	}

	// Holding fuel does not imply holding reports.
	fuelOnly := Identity{IsActive: true, HasRole: true, Permissions: map[string]ModulePermissions{
		"fuel": {All: true},
	}}
	if fuelOnly.Can("reports", http.MethodGet) {
		t.Fatal("a role without the reports module reached it")
	}
}
