package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ── helpers ────────────────────────────────────────────────────────────────

// clientIP extracts the remote IP address from the request.
// chi's chimiddleware.RealIP (applied at the router level) already rewrites
// r.RemoteAddr using the X-Real-IP / X-Forwarded-For headers, so we trust
// that value here. We deliberately do NOT re-read the proxy headers ourselves
// to prevent spoofed IP bypass attacks from untrusted clients.
func clientIP(r *http.Request) string {
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

func writeTooMany(w http.ResponseWriter, detail string, retryAfterSecs int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfterSecs))
	w.WriteHeader(http.StatusTooManyRequests)
	fmt.Fprintf(w,
		`{"title":"Too Many Requests","status":429,"detail":%q}`,
		detail,
	)
}

// ── responseCapture ────────────────────────────────────────────────────────

// responseCapture wraps http.ResponseWriter so the status code can be read
// after the downstream handler returns.
type responseCapture struct {
	http.ResponseWriter
	status int
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.status = code
	rc.ResponseWriter.WriteHeader(code)
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	if rc.status == 0 {
		rc.status = http.StatusOK
	}
	return rc.ResponseWriter.Write(b)
}

// Unwrap lets net/http middleware chains introspect the underlying writer.
func (rc *responseCapture) Unwrap() http.ResponseWriter { return rc.ResponseWriter }

// ── General per-IP rate limiter ────────────────────────────────────────────

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type ipRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*ipEntry
	r       rate.Limit
	burst   int
}

func newIPRateLimiter(r rate.Limit, burst int) *ipRateLimiter {
	l := &ipRateLimiter{
		entries: make(map[string]*ipEntry),
		r:       r,
		burst:   burst,
	}
	go l.cleanup()
	return l
}

func (l *ipRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.entries[ip]
	if !ok {
		e = &ipEntry{limiter: rate.NewLimiter(l.r, l.burst)}
		l.entries[ip] = e
	}
	e.lastSeen = time.Now()
	return e.limiter.Allow()
}

func (l *ipRateLimiter) cleanup() {
	for range time.Tick(5 * time.Minute) {
		l.mu.Lock()
		for ip, e := range l.entries {
			if time.Since(e.lastSeen) > 10*time.Minute {
				delete(l.entries, ip)
			}
		}
		l.mu.Unlock()
	}
}

// RateLimit returns a Chi middleware that allows at most requestsPerMinute
// requests per minute per client IP (token-bucket, burst = requestsPerMinute).
func RateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	lim := newIPRateLimiter(rate.Limit(float64(requestsPerMinute)/60.0), requestsPerMinute)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !lim.allow(clientIP(r)) {
				writeTooMany(w, "Rate limit exceeded. Try again in a moment.", 60)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ── Auth login attempt limiter ─────────────────────────────────────────────

type loginRecord struct {
	failures     int
	blockedUntil time.Time
	lastAttempt  time.Time
}

type loginAttemptStore struct {
	mu       sync.Mutex
	records  map[string]*loginRecord
	maxFails int
	blockFor time.Duration
}

func newLoginAttemptStore(maxFails int, blockFor time.Duration) *loginAttemptStore {
	s := &loginAttemptStore{
		records:  make(map[string]*loginRecord),
		maxFails: maxFails,
		blockFor: blockFor,
	}
	go s.cleanup()
	return s
}

func (s *loginAttemptStore) cleanup() {
	for range time.Tick(5 * time.Minute) {
		s.mu.Lock()
		for ip, rec := range s.records {
			if time.Now().After(rec.blockedUntil) && time.Since(rec.lastAttempt) > 10*time.Minute {
				delete(s.records, ip)
			}
		}
		s.mu.Unlock()
	}
}

func (s *loginAttemptStore) check(ip string) (blocked bool, retryAfter time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[ip]
	if !ok {
		return false, 0
	}
	if time.Now().Before(rec.blockedUntil) {
		return true, time.Until(rec.blockedUntil)
	}
	return false, 0
}

func (s *loginAttemptStore) onFailure(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[ip]
	if !ok {
		rec = &loginRecord{}
		s.records[ip] = rec
	}
	// Reset counter if previous block has already expired
	if time.Now().After(rec.blockedUntil) && rec.failures >= s.maxFails {
		rec.failures = 0
	}
	rec.failures++
	rec.lastAttempt = time.Now()
	if rec.failures >= s.maxFails {
		rec.blockedUntil = time.Now().Add(s.blockFor)
	}
}

func (s *loginAttemptStore) onSuccess(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, ip)
}

// LoginRateLimit returns a Chi middleware that tracks failed login attempts
// (HTTP 401 responses). After maxFails consecutive failures the IP is blocked
// for blockDuration. A successful login resets the counter.
func LoginRateLimit(maxFails int, blockDuration time.Duration) func(http.Handler) http.Handler {
	store := newLoginAttemptStore(maxFails, blockDuration)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)

			if blocked, wait := store.check(ip); blocked {
				secs := int(wait.Seconds()) + 1
				writeTooMany(w,
					fmt.Sprintf("Too many failed attempts. Try again in %d seconds.", secs),
					secs,
				)
				return
			}

			rc := &responseCapture{ResponseWriter: w}
			next.ServeHTTP(rc, r)

			switch rc.status {
			case http.StatusUnauthorized:
				store.onFailure(ip)
			case http.StatusOK, 0:
				store.onSuccess(ip)
			}
		})
	}
}
