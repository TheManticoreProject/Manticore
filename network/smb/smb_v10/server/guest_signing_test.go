package server

import (
	"net"
	"testing"

	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// TestClientDoesNotSignGuestSession logs an unknown user on to a server that
// offers signing and admits guests. The server reports SMB_SETUP_GUEST and does
// not sign, since a guest session has no key; [MS-SMB] 3.2.5.3 says the client
// MUST NOT activate signing then. A client that did refused the session, because
// the unsigned response failed its signature check.
func TestClientDoesNotSignGuestSession(t *testing.T) {
	config := conformanceConfig(SigningEnabled)
	config.AllowGuest = true
	_, transportEnd := pipedServer(t, config)
	client := smb1client.NewFromTransport(transportEnd, net.IPv4(127, 0, 0, 1), 445)
	if err := client.Negotiate(); err != nil {
		t.Fatalf("Negotiate() error = %v", err)
	}
	creds, err := credentials.NewCredentials(captureDomain, "nobody", capturePassword, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := client.SessionSetup(creds); err != nil {
		t.Fatalf("SessionSetup() as guest error = %v", err)
	}
	if client.Connection.IsSigningActive {
		t.Error("signing is active on a guest session")
	}
	if !client.Session.Guest {
		t.Error("the session does not report that it is a guest")
	}
	if _, err := client.Echo([]byte("guest")); err != nil {
		t.Errorf("Echo() on the guest session error = %v", err)
	}
}
