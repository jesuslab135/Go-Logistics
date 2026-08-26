package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"fleet/internal/auth"
	"fleet/internal/db/gen"
	"fleet/internal/domain/purchaseorder"
	"fleet/internal/platform/storage"
)

// newTestRouter builds the full route tree. Constructing it is itself the
// assertion for route conflicts: Gin panics on a duplicate or ambiguous path.
func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	return NewRouter(Deps{
		Queries: gen.New(nil),
		Tokens:  auth.NewTokenService("test-secret", "fleet", time.Hour, 24*time.Hour),
		Storage: storage.NewLocal(t.TempDir(), "/media"),
	})
}

func TestRouterRegistersRoutes(t *testing.T) {
	routes := make(map[string]bool)
	for _, r := range newTestRouter(t).Routes() {
		routes[r.Method+" "+r.Path] = true
	}

	for _, want := range []string{
		"POST /auth/login",
		"POST /auth/logout",
		"POST /auth/switch-company",
		"GET /api/v1/companies",
		"POST /api/v1/companies",
		"DELETE /api/v1/companies/:id",
		"GET /api/v1/assets",
		"POST /api/v1/employees/:id/set-password",
		"POST /api/v1/tire-assignment-requests/:id/approve",
		"GET /api/v1/asset-trailer-assignments",
		"GET /api/v1/fuel-entries",
		"GET /api/v1/tire-mount-logs",
		"POST /api/v1/notifications/:id/read",
		"GET /api/v1/part-inventory",
		"GET /api/v1/purchase-order-line-items",
		"POST /api/v1/uploads",
		"GET /api/v1/assets/:id/fuel-entries",
		"GET /api/v1/assets/:id/vehicle",
		"GET /api/v1/work-order-sub-line-items/:id/labor-entries",
		"GET /api/v1/me/permissions",
		"GET /api/v1/notifications",
		"GET /api/v1/dashboard/stats",
		"GET /api/v1/groups",
		"PUT /api/v1/groups/:id",
		"GET /api/v1/issues/:id/assigned-to",
		"POST /api/v1/issues/:id/assigned-to",
		"DELETE /api/v1/issues/:id/assigned-to/:employee_id",
		"GET /api/v1/issues/:id/watchers",
		"DELETE /api/v1/issues/:id/watchers/:employee_id",
		"GET /api/v1/work-orders/:id/issues",
		"DELETE /api/v1/work-orders/:id/issues/:issue_id",
		"GET /api/v1/work-orders/:id/faults",
		"DELETE /api/v1/work-orders/:id/faults/:fault_id",
		"GET /api/v1/service-entry-line-items/:id/issues",
		"POST /api/v1/work-order-line-items/:id/issues",
		"DELETE /api/v1/work-order-line-items/:id/issues/:issue_id",
		"DELETE /api/v1/service-entry-line-items/:id/issues/:issue_id",
		"GET /api/v1/admin/employees",
		"GET /api/v1/admin/employees/:id/companies",
		"PUT /api/v1/admin/employees/:id/companies",
		"GET /api/v1/admin/companies/:id/owner",
		"POST /api/v1/admin/companies/:id/set-owner",
		"GET /api/v1/admin/companies/:id/roles",
		"GET /api/v1/admin/companies/:id/work-order-statuses",
	} {
		if !routes[want] {
			t.Errorf("route %q not registered", want)
		}
	}
}

// Every /api/v1 route must reject an anonymous caller. A resource wired outside
// the authenticated groups would answer instead, which is how the pre-RBAC
// router leaked every module to any valid token.
// Switching companies lives outside /api/v1 but is still authenticated, so it
// is not covered by the sweep below.
func TestSwitchCompanyRequiresAuth(t *testing.T) {
	r := newTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/switch-company", strings.NewReader(`{"company_id":1}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAPIRoutesRequireAuth(t *testing.T) {
	r := newTestRouter(t)

	for _, route := range r.Routes() {
		if len(route.Path) < 8 || route.Path[:8] != "/api/v1/" {
			continue
		}
		req := httptest.NewRequest(route.Method, substituteParams(route.Path), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: got %d, want %d", route.Method, route.Path, w.Code, http.StatusUnauthorized)
		}
	}
}

func substituteParams(path string) string {
	out := ""
	for _, segment := range splitPath(path) {
		if segment == "" {
			continue
		}
		if segment[0] == ':' || segment[0] == '*' {
			segment = "1"
		}
		out += "/" + segment
	}
	return out
}

func splitPath(path string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			out = append(out, path[start:i])
			start = i + 1
		}
	}
	return out
}

// A retired write verb must stay routed. Leaving it unrouted returns 404, which
// a client cannot tell from a typo'd path; 405 with a stable code and the
// replacement in the message turns a bug report into a one-line fix.
func TestRetiredStatusLogWritesAnswer405(t *testing.T) {
	r := newTestRouter(t)

	writes := []struct{ method, path string }{
		{http.MethodPost, "/api/v1/work-orders/1/status-logs"},
		{http.MethodPut, "/api/v1/work-orders/1/status-logs/1"},
		{http.MethodDelete, "/api/v1/work-orders/1/status-logs/1"},
	}

	routes := make(map[string]bool)
	for _, route := range r.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	for _, w := range writes {
		path := strings.Replace(w.path, "/1/status-logs", "/:id/status-logs", 1)
		path = strings.Replace(path, "/status-logs/1", "/status-logs/:child_id", 1)
		if !routes[w.method+" "+path] {
			t.Errorf("%s %s is not routed; it must answer 405, not 404", w.method, path)
		}
	}
}

// The append-only history has no write routes of its own, so its request DTOs
// are gone. This pins that the read routes survive, since the whole point of
// retiring the writes was to keep the history readable and truthful.
func TestStatusLogReadsSurvive(t *testing.T) {
	routes := make(map[string]bool)
	for _, r := range newTestRouter(t).Routes() {
		routes[r.Method+" "+r.Path] = true
	}

	for _, want := range []string{
		"GET /api/v1/work-orders/:id/status-logs",
		"GET /api/v1/work-orders/:id/status-logs/:child_id",
	} {
		if !routes[want] {
			t.Errorf("missing route %q", want)
		}
	}
}

// The ledger's retired verbs must stay routed for the same reason the status
// log's do: a 404 reads as a wrong path, a 405 tells the client to reverse.
func TestInventoryJournalRoutes(t *testing.T) {
	routes := make(map[string]bool)
	for _, r := range newTestRouter(t).Routes() {
		routes[r.Method+" "+r.Path] = true
	}

	for _, want := range []string{
		"GET /api/v1/inventory-journal-entries",
		"POST /api/v1/inventory-journal-entries",
		"GET /api/v1/inventory-journal-entries/:id",
		"POST /api/v1/inventory-journal-entries/:id/reverse",
		// Retired, but routed: they answer 405 route_retired.
		"PUT /api/v1/inventory-journal-entries/:id",
		"DELETE /api/v1/inventory-journal-entries/:id",
	} {
		if !routes[want] {
			t.Errorf("missing route %q", want)
		}
	}
}

// Every transition the state machine defines must be routed, or an order can
// reach a state nothing moves it out of. Driven off the machine itself, so
// adding a transition without a route fails here.
func TestPurchaseOrderTransitionRoutes(t *testing.T) {
	routes := make(map[string]bool)
	for _, r := range newTestRouter(t).Routes() {
		routes[r.Method+" "+r.Path] = true
	}

	for _, action := range purchaseorder.Actions() {
		want := "POST /api/v1/purchase-orders/:id/" + action
		if !routes[want] {
			t.Errorf("missing route %q", want)
		}
	}

	for _, want := range []string{
		"GET /api/v1/purchase-orders/:id/status-logs",
		"GET /api/v1/purchase-orders/:id/status-logs/:child_id",
	} {
		if !routes[want] {
			t.Errorf("missing route %q", want)
		}
	}
}

// Totals are computed, so departing from the formula must go through the
// audited action rather than a writable field.
func TestTotalOverrideRoutes(t *testing.T) {
	routes := make(map[string]bool)
	for _, r := range newTestRouter(t).Routes() {
		routes[r.Method+" "+r.Path] = true
	}

	for _, want := range []string{
		"POST /api/v1/work-orders/:id/override-total",
		"POST /api/v1/purchase-orders/:id/override-total",
		"POST /api/v1/service-entries/:id/override-total",
	} {
		if !routes[want] {
			t.Errorf("missing route %q", want)
		}
	}
}

// A referenced record cannot be deleted, so every archivable resource needs the
// pair of routes that retire it instead. A missing one leaves that catalog with
// no way to retire anything.
func TestArchiveRoutes(t *testing.T) {
	routes := make(map[string]bool)
	for _, r := range newTestRouter(t).Routes() {
		routes[r.Method+" "+r.Path] = true
	}

	for _, resource := range []string{
		"/api/v1/assets",
		"/api/v1/parts",
		"/api/v1/vendors",
		"/api/v1/service-tasks",
		"/api/v1/inspection-forms",
	} {
		for _, action := range []string{"/archive", "/restore"} {
			want := "POST " + resource + "/:id" + action
			if !routes[want] {
				t.Errorf("missing route %q", want)
			}
		}
	}
}
