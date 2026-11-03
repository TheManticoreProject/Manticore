package server

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
)

// TestServerAcceptContextHook asserts the OnAcceptContext hook is handed the
// per-session acceptor during the real session-setup path, mirroring the SMB2
// coverage. This is the seam a relay uses to install a ChallengeProvider on the
// inbound exchange.
func TestServerAcceptContextHook(t *testing.T) {
	called := false
	_, addr := serverWithAccount(t, SigningDisabled, func(c *Config) {
		c.OnAcceptContext = func(accept *spnego.AcceptContext) {
			called = true
		}
	})

	establishSession(t, addr, captureDomain, captureUsername, capturePassword, false)

	if !called {
		t.Fatal("OnAcceptContext was not invoked during session setup")
	}
}
