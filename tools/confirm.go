package tools

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ConfirmationTTL is how long a minted confirmation token stays valid.
const ConfirmationTTL = 5 * time.Minute

// Sentinel errors returned by ConfirmationStore.Consume.
var (
	ErrTokenUnknown  = errors.New("confirmation token not recognized (already used, expired, or never issued)")
	ErrTokenExpired  = errors.New("confirmation token expired")
	ErrTokenMismatch = errors.New("confirmation token was issued for a different operation or target")
)

type pendingOp struct {
	operation string
	target    string
	expiry    time.Time
}

// ConfirmationStore backs the two-phase confirmation flow for create and delete
// operations. The first call mints a single-use, target-bound token with a short
// TTL; the second call consumes it to authorize exactly one execution. It is
// safe for concurrent use. State is per-process and resets on restart — the safe
// direction, since after a restart a destructive op must be re-previewed.
type ConfirmationStore struct {
	mu    sync.Mutex
	items map[string]pendingOp
	ttl   time.Duration
	now   func() time.Time // injectable clock; time.Now in production, fake in tests
}

// NewConfirmationStore returns an empty store using the default TTL and clock.
func NewConfirmationStore() *ConfirmationStore {
	return &ConfirmationStore{
		items: make(map[string]pendingOp),
		ttl:   ConfirmationTTL,
		now:   time.Now,
	}
}

// Mint issues a fresh single-use token authorizing operation on target. The
// token is bound to that exact (operation, target) pair and expires after the
// store's TTL. The only error path is a crypto/rand failure.
func (s *ConfirmationStore) Mint(operation, target string) (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("generating confirmation token: %w", err)
	}
	token := hex.EncodeToString(buf[:])

	s.mu.Lock()
	defer s.mu.Unlock()
	s.sweepExpiredLocked()
	s.items[token] = pendingOp{operation: operation, target: target, expiry: s.now().Add(s.ttl)}
	return token, nil
}

// Consume validates that token was minted for exactly this operation and target
// and has not expired, then deletes it (single-use). It returns a sentinel error
// (ErrTokenUnknown / ErrTokenExpired / ErrTokenMismatch) otherwise. A mismatch
// leaves the token in place so a legitimate retry with the correct target still
// works within the TTL.
func (s *ConfirmationStore) Consume(token, operation, target string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	op, ok := s.items[token]
	if !ok {
		return ErrTokenUnknown
	}
	if s.now().After(op.expiry) {
		delete(s.items, token) // evict dead entry
		return ErrTokenExpired
	}
	if op.operation != operation || op.target != target {
		return ErrTokenMismatch
	}
	delete(s.items, token) // single-use: consumed on success
	return nil
}

// sweepExpiredLocked drops expired entries to bound memory. The caller holds mu.
func (s *ConfirmationStore) sweepExpiredLocked() {
	now := s.now()
	for tok, op := range s.items {
		if now.After(op.expiry) {
			delete(s.items, tok)
		}
	}
}
