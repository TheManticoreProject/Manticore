package transport

import "time"

// SetNBTHandshakeTimeoutForTest shortens the NetBIOS session handshake deadline
// and returns a function restoring it.
func SetNBTHandshakeTimeoutForTest(d time.Duration) func() {
	previous := nbtHandshakeTimeout
	nbtHandshakeTimeout = d
	return func() { nbtHandshakeTimeout = previous }
}
