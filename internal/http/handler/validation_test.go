package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"fleet/internal/platform/apierr"
)

func bindBody(t *testing.T, body string, obj any) error {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return bindJSONValidated(c, obj)
}

func TestBindJSONValidatedReportsFieldDetails(t *testing.T) {
	var req struct {
		Password string `json:"password" binding:"required,min=8"`
	}

	err := bindBody(t, `{"password":"short"}`, &req)

	var ae *apierr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("want *apierr.Error, got %v", err)
	}
	if ae.Status != http.StatusUnprocessableEntity {
		t.Fatalf("got status %d, want 422", ae.Status)
	}
	details, ok := ae.Details.(map[string]string)
	if !ok {
		t.Fatalf("want map[string]string details, got %#v", ae.Details)
	}
	if details["password"] == "" {
		t.Fatalf("want a detail keyed by the json name, got %#v", details)
	}
}

func TestBindJSONValidatedAcceptsValidBody(t *testing.T) {
	var req struct {
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := bindBody(t, `{"password":"longenough"}`, &req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Password != "longenough" {
		t.Fatalf("got %q", req.Password)
	}
}

func TestBindJSONValidatedRejectsMalformedJSON(t *testing.T) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}

	err := bindBody(t, `{`, &req)

	var ae *apierr.Error
	if !errors.As(err, &ae) || ae.Status != http.StatusBadRequest {
		t.Fatalf("want 400 apierr, got %v", err)
	}
}
