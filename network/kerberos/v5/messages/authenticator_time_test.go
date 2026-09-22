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
			ctime, cusec := NextAuthenticatorTimestamp(now)
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
