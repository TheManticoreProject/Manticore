package server

import "testing"

// TestNewSessionIDIsRandomNotSequential checks that SessionIds are independent
// random values, not a per-process counter. The previous scheme seeded a counter
// once and added 1 per session, so SessionIds were consecutive across every
// connection from one server — a fingerprint that also leaks a session count. A
// fresh random id per session removes both.
func TestNewSessionIDIsRandomNotSequential(t *testing.T) {
	s, err := NewServer(Config{})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	const n = 256
	seen := make(map[uint64]struct{}, n)
	var prev uint64
	for i := 0; i < n; i++ {
		id := s.newSessionID()
		if id == 0 {
			t.Fatal("newSessionID returned 0")
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("newSessionID returned a duplicate: %#x", id)
		}
		seen[id] = struct{}{}
		if i > 0 && id == prev+1 {
			t.Fatalf("SessionIds are consecutive (id %#x follows %#x); expected independent random values", id, prev)
		}
		prev = id
	}
}
