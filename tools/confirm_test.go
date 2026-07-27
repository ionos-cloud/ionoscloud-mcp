package tools

import (
	"sync"
	"testing"
	"time"
)

func TestConfirmationMintConsume(t *testing.T) {
	s := NewConfirmationStore()
	tok, err := s.Mint("delete_datacenter", "dc-1")
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if tok == "" {
		t.Fatal("Mint returned an empty token")
	}
	if err := s.Consume(tok, "delete_datacenter", "dc-1"); err != nil {
		t.Fatalf("Consume: %v", err)
	}
	// Single-use: the second Consume must fail.
	if err := s.Consume(tok, "delete_datacenter", "dc-1"); err != ErrTokenUnknown {
		t.Errorf("second Consume err = %v, want ErrTokenUnknown", err)
	}
}

func TestConfirmationTargetMismatch(t *testing.T) {
	s := NewConfirmationStore()
	tok, _ := s.Mint("delete_datacenter", "dc-A")
	// A token for DC-A must not delete DC-B; the token survives for the right target.
	if err := s.Consume(tok, "delete_datacenter", "dc-B"); err != ErrTokenMismatch {
		t.Errorf("Consume wrong target err = %v, want ErrTokenMismatch", err)
	}
	if err := s.Consume(tok, "delete_datacenter", "dc-A"); err != nil {
		t.Errorf("Consume correct target after mismatch err = %v, want nil", err)
	}
}

func TestConfirmationOperationMismatch(t *testing.T) {
	s := NewConfirmationStore()
	tok, _ := s.Mint("create_datacenter", "n|de/fra")
	if err := s.Consume(tok, "delete_datacenter", "n|de/fra"); err != ErrTokenMismatch {
		t.Errorf("Consume wrong operation err = %v, want ErrTokenMismatch", err)
	}
}

func TestConfirmationExpiry(t *testing.T) {
	now := time.Unix(1000, 0)
	s := &ConfirmationStore{items: map[string]pendingOp{}, ttl: ConfirmationTTL, now: func() time.Time { return now }}
	tok, _ := s.Mint("delete_datacenter", "dc-1")
	now = now.Add(ConfirmationTTL + time.Second) // advance past the TTL
	if err := s.Consume(tok, "delete_datacenter", "dc-1"); err != ErrTokenExpired {
		t.Errorf("Consume expired err = %v, want ErrTokenExpired", err)
	}
	// Expired entry is evicted, so a further Consume is unknown.
	if err := s.Consume(tok, "delete_datacenter", "dc-1"); err != ErrTokenUnknown {
		t.Errorf("Consume after expiry err = %v, want ErrTokenUnknown", err)
	}
}

func TestConfirmationUnknownToken(t *testing.T) {
	s := NewConfirmationStore()
	if err := s.Consume("nope", "delete_datacenter", "dc-1"); err != ErrTokenUnknown {
		t.Errorf("Consume unknown err = %v, want ErrTokenUnknown", err)
	}
}

func TestConfirmationConcurrentConsume(t *testing.T) {
	s := NewConfirmationStore()
	tok, _ := s.Mint("delete_datacenter", "dc-1")
	const n = 20
	var wg sync.WaitGroup
	var mu sync.Mutex
	winners := 0
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			if err := s.Consume(tok, "delete_datacenter", "dc-1"); err == nil {
				mu.Lock()
				winners++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if winners != 1 {
		t.Errorf("concurrent Consume winners = %d, want exactly 1", winners)
	}
}

func TestConfirmationTokensDistinct(t *testing.T) {
	s := NewConfirmationStore()
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		tok, err := s.Mint("create_datacenter", "n|de/fra")
		if err != nil {
			t.Fatalf("Mint: %v", err)
		}
		if seen[tok] {
			t.Fatalf("duplicate token %q", tok)
		}
		seen[tok] = true
	}
}
