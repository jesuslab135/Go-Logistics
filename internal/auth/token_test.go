package auth

import (
	"testing"
	"time"
)

func TestRefreshTokenCarriesUniqueJTI(t *testing.T) {
	s := NewTokenService("secret", "fleet", time.Hour, 24*time.Hour)

	first, err := s.Issue(7, 3, false)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	second, err := s.Issue(7, 3, false)
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

	pair, err := s.Issue(7, 3, false)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if _, err := s.ParseRefresh(pair.AccessToken); err == nil {
		t.Fatal("access token accepted as a refresh token")
	}
}
