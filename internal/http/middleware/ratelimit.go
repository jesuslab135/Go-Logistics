package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"fleet/internal/platform/apierr"
)

// Rate limiting.
//
// /auth/login was an unthrottled guessing oracle: no lockout, no backoff, and
// nothing at the edge either. Each attempt also costs a bcrypt compare, so the
// same flood that guesses passwords is simultaneously a CPU-exhaustion attack —
// the work is done before the password is known to be wrong, which is the whole
// point of a slow hash.
//
// The limiter is in-process, and therefore per-replica: with N replicas the
// effective limit is N x rate. That is the right trade for a single-container
// deployment, and Allow is where a shared store would slot in if that stops
// being true. nginx carries a second, coarser limit at the edge
// (deploy/nginx/go-logistics.conf), so a flood that never reaches Go is still
// bounded.

// Limiter is a token bucket per key, refilling continuously at rate/per.
//
// Continuous refill rather than a fixed window: a window resets on a boundary,
// so a caller refused at the start of one waits the whole window, and two
// callers arriving either side of a boundary get twice the intended rate
// between them.
type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64       // tokens per second
	burst   float64       // bucket capacity
	window  time.Duration // the period rate was expressed over
	now     func() time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

// NewLimiter allows rate requests per per, for each key, bursting to rate.
func NewLimiter(rate int, per time.Duration) *Limiter {
	if rate < 1 {
		rate = 1
	}
	if per <= 0 {
		per = time.Minute
	}
	return &Limiter{
		buckets: make(map[string]*bucket),
		rate:    float64(rate) / per.Seconds(),
		burst:   float64(rate),
		window:  per,
		now:     time.Now,
	}
}

// Allow consumes a token for key, reporting whether one was available and, when
// it was not, how long until one is.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	}

	b.tokens = min(l.burst, b.tokens+now.Sub(b.last).Seconds()*l.rate)
	b.last = now

	if b.tokens < 1 {
		return false, time.Duration((1 - b.tokens) / l.rate * float64(time.Second))
	}
	b.tokens--
	return true, 0
}

// reap drops buckets untouched for two windows. Without it the map grows by one
// entry per distinct client address the process has ever seen.
func (l *Limiter) reap() {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := l.now().Add(-2 * l.window)
	for k, b := range l.buckets {
		if b.last.Before(cutoff) {
			delete(l.buckets, k)
		}
	}
}

// size is the number of live buckets. For tests.
func (l *Limiter) size() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}

// StartReaper runs reap on a ticker for the life of the process. Called once at
// wiring time; there is nothing to stop, because the limiter lives as long as
// the router does.
func (l *Limiter) StartReaper(every time.Duration) {
	go func() {
		for range time.Tick(every) {
			l.reap()
		}
	}()
}

// RateLimit refuses a request whose key has spent its budget.
//
// An empty key means the request carries nothing to attribute the cost to. It
// is allowed through rather than collapsed into one shared bucket, which would
// turn a missing key into an outage for everyone at once.
func RateLimit(l *Limiter, keyFn func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFn(c)
		if key == "" {
			c.Next()
			return
		}
		ok, retryAfter := l.Allow(key)
		if !ok {
			// Rounded up: a Retry-After the caller can honour literally and
			// still be served, rather than one that rounds down into a second
			// refusal.
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))
			apierr.Abort(c, apierr.New(http.StatusTooManyRequests, "rate_limited",
				"too many requests; retry shortly"))
			return
		}
		c.Next()
	}
}

// ByIP keys on the client address, for routes reached before authentication.
//
// Note that ClientIP is only as trustworthy as the proxy configuration behind
// it: gin.SetTrustedProxies is not called anywhere yet, so X-Forwarded-For is
// caller-controlled and this key can be rotated at will. The nginx limit keys
// on the real peer address, which is why both layers exist.
func ByIP(c *gin.Context) string { return "ip:" + c.ClientIP() }

// ByEmployee keys on the authenticated employee, so one account cannot spend
// the whole API budget from many addresses. It falls back to the address when
// no claims are present, so an unauthenticated request is still bounded.
func ByEmployee(c *gin.Context) string {
	if claims, ok := ClaimsOf(c); ok {
		return "emp:" + strconv.FormatInt(claims.EmployeeID(), 10)
	}
	return ByIP(c)
}
