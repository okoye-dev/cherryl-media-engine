package handlers

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// loginLimiter slows down password guessing. It counts failed sign-ins for an email address
// and for a network address, and refuses further tries for a while once either has too many.
type loginLimiter struct {
	mu       sync.Mutex
	failures map[string][]time.Time
	max      int
	window   time.Duration
}

func newLoginLimiter(max int, window time.Duration) *loginLimiter {
	return &loginLimiter{failures: map[string][]time.Time{}, max: max, window: window}
}

func (l *loginLimiter) recent(key string, now time.Time) []time.Time {
	kept := l.failures[key][:0]
	for _, t := range l.failures[key] {
		if now.Sub(t) < l.window {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		delete(l.failures, key)
	} else {
		l.failures[key] = kept
	}
	return kept
}

// blocked reports whether any of the keys has used up its tries.
func (l *loginLimiter) blocked(keys ...string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	for _, k := range keys {
		if len(l.recent(k, now)) >= l.max {
			return true
		}
	}
	return false
}

func (l *loginLimiter) fail(keys ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	for _, k := range keys {
		l.failures[k] = append(l.recent(k, now), now)
	}
}

func (l *loginLimiter) clear(keys ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, k := range keys {
		delete(l.failures, k)
	}
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
