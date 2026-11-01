package messages

import (
	"sync"
	"time"
)

var authenticatorClock struct {
	sync.Mutex
	lastMicros int64
}

// NextAuthenticatorTimestamp returns a Kerberos ctime/cusec pair that is
// unique in the process for a given offset: now, advanced one microsecond when
// the clock has not moved since the last authenticator the process created,
// plus offset. RFC 4120 replay caches key authenticators by this pair.
// Uniqueness is kept on the local clock and the offset (the correction learned
// from a peer's KRB_AP_ERR_SKEW) is added after, so an authenticator for a peer
// whose clock is behind is not pulled forward to the latest authenticator
// created for a peer whose clock is not.
func NextAuthenticatorTimestamp(now time.Time, offset time.Duration) (time.Time, int) {
	authenticatorClock.Lock()
	defer authenticatorClock.Unlock()

	micros := now.UTC().UnixMicro()
	if micros <= authenticatorClock.lastMicros {
		micros = authenticatorClock.lastMicros + 1
	}
	authenticatorClock.lastMicros = micros
	unique := time.UnixMicro(micros).UTC().Add(offset)
	return unique, unique.Nanosecond() / 1000
}
