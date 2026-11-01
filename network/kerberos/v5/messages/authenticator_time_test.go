package messages

import (
	"sync"
	"testing"
	"time"
)

func TestNextAuthenticatorTimestampConcurrentUniqueness(t *testing.T) {
	const count = 256
	now := time.Now().UTC()
	pairs := make(chan int64, count)
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctime, cusec := NextAuthenticatorTimestamp(now, 0)
			pairs <- ctime.Unix()*1_000_000 + int64(cusec)
		}()
	}
	wg.Wait()
	close(pairs)

	seen := make(map[int64]bool, count)
	for pair := range pairs {
		if seen[pair] {
			t.Fatalf("duplicate authenticator timestamp %d", pair)
		}
		seen[pair] = true
	}
}

// TestNextAuthenticatorTimestampKeepsNegativeOffset checks that an authenticator
// for a peer whose clock is behind ours keeps that peer's time after the process
// has stamped an authenticator for a peer on our own time. Clamping after the
// offset is added would pull it forward to our clock, which the lagging peer
// rejects with KRB_AP_ERR_SKEW.
func TestNextAuthenticatorTimestampKeepsNegativeOffset(t *testing.T) {
	const offset = -12 * time.Hour
	now := time.Now().UTC()
	first, _ := NextAuthenticatorTimestamp(now, 0)
	second, cusec := NextAuthenticatorTimestamp(now, offset)
	if d := second.Sub(now.Add(offset)); d < 0 || d > time.Second {
		t.Fatalf("ctime with offset %v is %v from the offset clock, want within a second", offset, d)
	}
	if second.Nanosecond()/1000 != cusec {
		t.Errorf("cusec = %d, want %d", cusec, second.Nanosecond()/1000)
	}
	// Uniqueness is still kept on the local clock.
	if !second.Add(-offset).After(first) {
		t.Errorf("second authenticator's local time %v is not after the first %v", second.Add(-offset), first)
	}
}
