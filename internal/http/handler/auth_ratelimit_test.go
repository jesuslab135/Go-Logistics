package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"fleet/internal/auth"
)

// refusingVerifier fails every credential, like an attacker guessing.
type refusingVerifier struct{ calls int }

func (v *refusingVerifier) Verify(context.Context, string, string) (Identity, error) {
	v.calls++
	return Identity{}, ErrInvalidCredentials
}

// Guessing one account must be bounded per account, not merely per address: an
// attacker has many addresses and only one target. The per-IP limit on the
// route cannot express that, and is deliberately loose so a NAT'd office does
// not lock itself out.
func TestLoginIsRateLimitedPerAccountAcrossAddresses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	verifier := &refusingVerifier{}
	h := NewAuthHandler(auth.NewTokenService("s", "i", 0, 0), verifier, nil)

	r := gin.New()
	r.POST("/auth/login", h.Login)

	attempt := func(email, fromIP string) int {
		body, _ := json.Marshal(map[string]string{"email": email, "password": "guess"})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = fromIP + ":1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	// Ten guesses at one account, each from a different address.
	for i := range 10 {
		if code := attempt("victim@example.test", ipFor(i)); code != http.StatusUnauthorized {
			t.Fatalf("attempt %d = %d, want 401", i+1, code)
		}
	}

	// The eleventh is refused on the account, not the address.
	if code := attempt("victim@example.test", ipFor(99)); code != http.StatusTooManyRequests {
		t.Errorf("11th attempt from a fresh address = %d, want 429", code)
	}

	// The expensive half must not have run for the refused attempt.
	if verifier.calls != 10 {
		t.Errorf("verifier ran %d times, want 10: a refused attempt must not reach bcrypt", verifier.calls)
	}

	// A different account is unaffected — the limit is per account, not global.
	if code := attempt("someone.else@example.test", ipFor(0)); code != http.StatusUnauthorized {
		t.Errorf("a different account = %d, want 401: one account's budget must not exhaust another's", code)
	}
}

// The key is the account, so a different case must not mint a fresh budget.
// (Surrounding whitespace cannot reach the limiter: the email binding rejects
// it as malformed first, which is a 400 before any budget is consulted.)
func TestLoginAttemptKeyIsCaseInsensitive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAuthHandler(auth.NewTokenService("s", "i", 0, 0), &refusingVerifier{}, nil)
	r := gin.New()
	r.POST("/auth/login", h.Login)

	attempt := func(email string) int {
		body, _ := json.Marshal(map[string]string{"email": email, "password": "guess"})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	for range 10 {
		attempt("victim@example.test")
	}
	if code := attempt("VICTIM@Example.TEST"); code != http.StatusTooManyRequests {
		t.Errorf("same account in different case = %d, want 429", code)
	}
}

func ipFor(n int) string {
	return "203.0.113." + string(rune('0'+n%10))
}
