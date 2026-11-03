package server_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
)

// TestServerAcceptContextHook asserts the OnAcceptContext hook is handed the
// per-session acceptor during the real session-setup path, and that leaving the
// default behaviour untouched does not disturb a normal logon. This is the seam
// a relay uses to install a ChallengeProvider on the inbound exchange.
func TestServerAcceptContextHook(t *testing.T) {
	called := false
	_, port := startServer(t, server.Config{
		OnAcceptContext: func(accept *spnego.AcceptContext) {
			called = true
		},
	})

	if _, err := dial(t, port, dialects.SMB2_DIALECT_2_1_0, "Passw0rd!"); err != nil {
		t.Fatalf("dial: %v", err)
	}
	if !called {
		t.Fatal("OnAcceptContext was not invoked during session setup")
	}
}
