package auth

import (
	"testing"
	"time"
)

func i64(v int64) *int64 { return &v }

// NewTokenService takes the secret as a string (it converts internally) —
// verified against internal/auth/token.go:50.
func newTestService(t *testing.T) *TokenService {
	t.Helper()
	return NewTokenService("test-secret-value", "fleet-test", time.Minute, time.Hour)
}

func TestRefreshTokenCarriesUniqueJTI(t *testing.T) {
	s := NewTokenService("secret", "fleet", time.Hour, 24*time.Hour)

	first, err := s.Issue(7, i64(3), nil, false)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	second, err := s.Issue(7, i64(3), nil, false)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	firstClaims, err := s.ParseRefresh(first.RefreshToken)
	if err != nil {
		t.Fatalf("parse refresh: %v", err)
	}
	secondClaims, err := s.ParseRefresh(second.RefreshToken)
	if err != nil {
		t.Fatalf("parse refresh: %v", err)
	}

	if firstClaims.ID == "" {
		t.Fatal("refresh token carries no jti; it cannot be revoked")
	}
	if firstClaims.ID == secondClaims.ID {
		t.Fatalf("jti must be unique per issue, got %q twice", firstClaims.ID)
	}
}

// Access tokens are never revoked, so they carry no jti and must not be
// accepted where a refresh token is expected.
func TestParseRefreshRejectsAccessToken(t *testing.T) {
	s := NewTokenService("secret", "fleet", time.Hour, 24*time.Hour)

	pair, err := s.Issue(7, i64(3), nil, false)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if _, err := s.ParseRefresh(pair.AccessToken); err == nil {
		t.Fatal("access token accepted as a refresh token")
	}
}

func TestIssueAndParseCarriesCompanyAndAccount(t *testing.T) {
	s := newTestService(t)
	pair, err := s.Issue(42, i64(7), i64(3), true)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	claims, err := s.Parse(pair.AccessToken, TypeAccess)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.EmployeeID() != 42 {
		t.Fatalf("employee id = %d, want 42", claims.EmployeeID())
	}
	if claims.CompanyID == nil || *claims.CompanyID != 7 {
		t.Fatalf("company id = %v, want 7", claims.CompanyID)
	}
	if claims.AccountID == nil || *claims.AccountID != 3 {
		t.Fatalf("account id = %v, want 3", claims.AccountID)
	}
}

// A company-less session is the whole point of this change: an account owner who
// has not created a company yet must still receive a usable token. The claim is
// ABSENT rather than zero — a zero would flow into a scoped query as a real
// argument and silently scope to a company that cannot exist.
func TestIssueOmitsCompanyWhenThereIsNone(t *testing.T) {
	s := newTestService(t)
	pair, err := s.Issue(42, nil, i64(3), false)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	claims, err := s.Parse(pair.AccessToken, TypeAccess)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.CompanyID != nil {
		t.Fatalf("company id = %v, want nil", *claims.CompanyID)
	}
	if claims.AccountID == nil || *claims.AccountID != 3 {
		t.Fatalf("account id = %v, want 3", claims.AccountID)
	}
}
