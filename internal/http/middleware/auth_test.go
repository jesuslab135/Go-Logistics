package middleware

import (
	"context"
	"testing"

	"fleet/internal/auth"
)

// authContext is what Auth attaches to every request's context. This drives
// the real producer (authContext) into the real consumer (CompanyFromContext)
// so a producer/consumer type mismatch fails a test rather than silently
// scoping every company-scoped query to 0. That is exactly what shipped here
// once: the company id was stored as *int64 but CompanyFromContext asserted
// int64, so the assertion always failed and every request read company 0.
func TestAuthContextRoundTripsCompany(t *testing.T) {
	company := int64(42)
	claims := &auth.Claims{CompanyID: &company}
	claims.Subject = "7"

	ctx := authContext(context.Background(), claims)

	if got := CompanyFromContext(ctx); got != company {
		t.Fatalf("CompanyFromContext = %d, want %d", got, company)
	}
	if got := EmployeeFromContext(ctx); got != 7 {
		t.Fatalf("EmployeeFromContext = %d, want 7", got)
	}
}

// A company-less session (an account owner with no company yet) has no tenant
// to scope to. It must read back as 0 — CompanyFromContext's "absent" value —
// not panic and not be silently attributed to some other tenant.
func TestAuthContextCompanylessSessionReadsZero(t *testing.T) {
	claims := &auth.Claims{}
	claims.Subject = "9"

	ctx := authContext(context.Background(), claims)

	if got := CompanyFromContext(ctx); got != 0 {
		t.Fatalf("CompanyFromContext = %d, want 0", got)
	}
	if got := EmployeeFromContext(ctx); got != 9 {
		t.Fatalf("EmployeeFromContext = %d, want 9", got)
	}
}
