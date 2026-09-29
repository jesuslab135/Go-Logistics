package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestLimiterAllowsUpToTheRateThenRefuses(t *testing.T) {
	l := NewLimiter(3, time.Minute)
	for i := range 3 {
		if ok, _ := l.Allow("a"); !ok {
			t.Fatalf("request %d of 3 should have been allowed", i+1)
		}
	}
	ok, retryAfter := l.Allow("a")
	if ok {
		t.Error("the 4th request in a 3-per-minute window should have been refused")
	}
	if retryAfter <= 0 || retryAfter > time.Minute {
		t.Errorf("retryAfter = %v, want a positive duration no longer than the window", retryAfter)
	}
}

func TestLimiterKeysAreIndependent(t *testing.T) {
	l := NewLimiter(1, time.Minute)
	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("first request for a should be allowed")
	}
	if ok, _ := l.Allow("b"); !ok {
		t.Error("b must not be refused because a spent its budget")
	}
}

// The bucket refills continuously rather than resetting on a boundary, so a
// caller that waits out its own refusal is served rather than being made to
// wait for an arbitrary window edge.
func TestLimiterRefillsOverTime(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewLimiter(1, time.Second)
	l.now = func() time.Time { return now }

	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("first request should be allowed")
	}
	if ok, _ := l.Allow("a"); ok {
		t.Fatal("second request in the same window should be refused")
	}

	now = now.Add(time.Second)
	if ok, _ := l.Allow("a"); !ok {
		t.Error("the budget should have refilled a second later")
	}
}

// A bucket per client IP is a per-request allocation if nothing ever removes
// them, which is a slow leak on a long-lived process.
func TestLimiterReapsIdleBuckets(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewLimiter(1, time.Minute)
	l.now = func() time.Time { return now }

	l.Allow("a")
	if got := l.size(); got != 1 {
		t.Fatalf("bucket count = %d, want 1", got)
	}

	now = now.Add(10 * time.Minute)
	l.reap()
	if got := l.size(); got != 0 {
		t.Errorf("bucket count after reaping = %d, want 0", got)
	}
}

// An exhausted caller gets 429 and is told when to come back, rather than a
// bare refusal it can only respond to by retrying immediately.
func TestRateLimitMiddlewareRefusesWith429AndRetryAfter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(NewLimiter(1, time.Minute), ByIP))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	first := httptest.NewRecorder()
	r.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/x", nil))
	if first.Code != http.StatusOK {
		t.Fatalf("first request = %d, want 200", first.Code)
	}

	second := httptest.NewRecorder()
	r.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/x", nil))
	if second.Code != http.StatusTooManyRequests {
		t.Errorf("second request = %d, want 429", second.Code)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Error("a 429 must carry Retry-After, or the client can only guess")
	}
}

// A limiter whose key function yields nothing must not collapse every caller
// into one shared bucket — that would turn a missing key into a global outage.
func TestRateLimitSkipsWhenThereIsNoKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(NewLimiter(1, time.Minute), func(*gin.Context) string { return "" }))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := range 3 {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("request %d = %d, want 200: an absent key must not rate-limit", i+1, w.Code)
		}
	}
}

// Concurrent callers must not be able to overspend a bucket between the read
// and the write. This does not need -race to be useful: if the accounting were
// not held under the lock, the allowed count would exceed the budget.
func TestLimiterDoesNotOverspendUnderConcurrency(t *testing.T) {
	const budget = 50
	const callers = 200

	l := NewLimiter(budget, time.Hour) // long window: no refill during the test

	var wg sync.WaitGroup
	results := make(chan bool, callers)
	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, _ := l.Allow("shared")
			results <- ok
		}()
	}
	wg.Wait()
	close(results)

	allowed := 0
	for ok := range results {
		if ok {
			allowed++
		}
	}
	if allowed != budget {
		t.Errorf("allowed %d of %d concurrent requests, want exactly %d", allowed, callers, budget)
	}
}
