package auth

import "sync"

// replayGuard makes a login code single-use per account, per PRD 016. It
// tracks, for each account that has ever completed a login, only the most
// recently used TOTP step — one entry per account, never per code or per
// attempt, so it is bounded by the number of accounts and self-overwrites
// on every new successful login rather than growing without limit.
type replayGuard struct {
	mu       sync.Mutex
	usedStep map[string]uint64
}

// newReplayGuard builds an empty, ready-to-use replayGuard.
func newReplayGuard() *replayGuard {
	return &replayGuard{usedStep: make(map[string]uint64)}
}

// consume reports whether step has already completed a login for accountID.
// If not, it records step as used for accountID and returns false, so the
// caller may proceed; a second call with the same accountID and step always
// returns true.
func (g *replayGuard) consume(accountID string, step uint64) (alreadyUsed bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if used, ok := g.usedStep[accountID]; ok && used == step {
		return true
	}
	g.usedStep[accountID] = step
	return false
}
