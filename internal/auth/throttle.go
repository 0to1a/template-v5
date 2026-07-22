package auth

import (
	"sync"
	"time"
)

const (
	// loginFailureThreshold is how many failed SubmitLogin attempts for one
	// account within loginFailureWindow trigger a lockout.
	loginFailureThreshold = 5
	// loginFailureWindow is the sliding window failures are counted within.
	loginFailureWindow = 5 * time.Minute
	// loginLockoutDuration is how long an account stays locked out once
	// loginFailureThreshold is reached.
	loginLockoutDuration = 15 * time.Minute
)

// throttleEntry is one account's failure-tracking state.
type throttleEntry struct {
	failureCount int
	windowStart  time.Time
	lockedUntil  time.Time
}

// loginThrottle counts failed login attempts per account and locks an
// account out once too many land within one window. It is keyed by account
// (PublicUUID), never by the attacker-controlled email on the request, and
// is populated only after an account is known to exist — so its size is
// bounded by the number of real accounts with recent failures, not by
// attacker-controlled input. A successful login or an expired window prunes
// an account's entry, so it never grows without bound.
type loginThrottle struct {
	mu      sync.Mutex
	entries map[string]*throttleEntry
}

// newLoginThrottle builds an empty, ready-to-use loginThrottle.
func newLoginThrottle() *loginThrottle {
	return &loginThrottle{entries: make(map[string]*throttleEntry)}
}

// locked reports whether accountID is currently locked out at time now.
func (t *loginThrottle) locked(accountID string, now time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	entry, ok := t.entries[accountID]
	if !ok {
		return false
	}
	return now.Before(entry.lockedUntil)
}

// recordFailure counts one failed attempt for accountID at time now,
// locking the account out once loginFailureThreshold failures have landed
// within loginFailureWindow. A failure outside the current window starts a
// fresh window instead of accumulating indefinitely.
func (t *loginThrottle) recordFailure(accountID string, now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	entry, ok := t.entries[accountID]
	if !ok || now.Sub(entry.windowStart) > loginFailureWindow {
		entry = &throttleEntry{windowStart: now}
		t.entries[accountID] = entry
	}

	entry.failureCount++
	if entry.failureCount >= loginFailureThreshold {
		entry.lockedUntil = now.Add(loginLockoutDuration)
	}
}

// reset clears accountID's failure state after a successful login.
func (t *loginThrottle) reset(accountID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, accountID)
}
