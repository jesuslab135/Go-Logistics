package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func corsResponse(t *testing.T, origins []string, method, origin string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORS(origins))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(method, "/x", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if method == http.MethodOptions {
		req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// The deployed allowlist held only the backend's own origin, which no browser
// ever sends, so every client got a response with no Allow-Origin header at all.
func TestCORSAllowOrigin(t *testing.T) {
	tests := []struct {
		name    string
		origins []string
		method  string
		origin  string
		want    string
	}{
		{
			name:    "an allowlisted origin is echoed",
			origins: []string{"https://app.example.com", "http://localhost:5173"},
			method:  http.MethodGet,
			origin:  "http://localhost:5173",
			want:    "http://localhost:5173",
		},
		{
			name:    "a preflight from an allowlisted origin is echoed",
			origins: []string{"http://localhost:5173"},
			method:  http.MethodOptions,
			origin:  "http://localhost:5173",
			want:    "http://localhost:5173",
		},
		{
			name:    "an origin outside the allowlist gets no header",
			origins: []string{"https://app.example.com"},
			method:  http.MethodGet,
			origin:  "https://evil.example.com",
			want:    "",
		},
		{
			name:    "a wildcard allowlist allows any origin",
			origins: []string{"*"},
			method:  http.MethodGet,
			origin:  "https://anything.example.com",
			want:    "*",
		},
		{
			name:    "an allowlist holding only the backend's own origin serves no browser",
			origins: []string{"https://go-logistics.jesuslab135.com"},
			method:  http.MethodOptions,
			origin:  "http://localhost:5173",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := corsResponse(t, tt.origins, tt.method, tt.origin)
			if got := w.Header().Get("Access-Control-Allow-Origin"); got != tt.want {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCORSPreflightIsNoContent(t *testing.T) {
	w := corsResponse(t, []string{"http://localhost:5173"}, http.MethodOptions, "http://localhost:5173")
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("preflight returned no Access-Control-Allow-Methods")
	}
}
