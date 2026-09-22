package messages

import (
	"sync"
	"time"
)

var authenticatorClock struct {
	sync.Mutex
	lastMicros int64
}

// NextAuthenticatorTimestamp returns a process-unique Kerberos ctime/cusec
// pair. RFC 4120 replay caches key authenticators by this pair, so callers must
// advance one microsecond when the wall clock has not advanced since the last
// authenticator created by the process.
func NextAuthenticatorTimestamp(now time.Time) (time.Time, int) {
	authenticatorClock.Lock()
	defer authenticatorClock.Unlock()

	micros := now.UTC().UnixMicro()
	if micros <= authenticatorClock.lastMicros {
		micros = authenticatorClock.lastMicros + 1
	}
	authenticatorClock.lastMicros = micros
	unique := time.UnixMicro(micros).UTC()
	return unique, unique.Nanosecond() / 1000
}
