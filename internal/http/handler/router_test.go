package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"fleet/internal/auth"
	"fleet/internal/db/gen"
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
		"GET /api/v1/companies",
		"POST /api/v1/companies",
		"DELETE /api/v1/companies/:id",
		"GET /api/v1/assets",
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
	} {
		if !routes[want] {
			t.Errorf("route %q not registered", want)
		}
	}
}

// Every /api/v1 route must reject an anonymous caller. A resource wired outside
// the authenticated groups would answer instead, which is how the pre-RBAC
// router leaked every module to any valid token.
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
