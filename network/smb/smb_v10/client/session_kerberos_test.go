package client

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// TestSessionSetupKerberosGuards covers the input-validation guards that return
// before any network/transport access. The full single-leg handshake requires a
// live KDC + SMB1 server and is covered by an integration test, not here.
func TestSessionSetupKerberosGuards(t *testing.T) {
	c := &Client{}

	if err := c.SessionSetupKerberos(nil, "dc.corp.local", "cifs/host.corp.local"); err == nil {
		t.Error("expected an error for nil credentials")
	} else if !strings.Contains(err.Error(), "credentials") {
		t.Errorf("unexpected error for nil credentials: %v", err)
	}

	creds, err := credentials.NewCredentials("CORP", "user", "Passw0rd!", "")
	if err != nil {
		t.Fatalf("NewCredentials: %s", err)
	}
	if err := c.SessionSetupKerberos(creds, "", "cifs/host.corp.local"); err == nil {
		t.Error("expected an error for an empty KDC host")
	} else if !strings.Contains(err.Error(), "KDC host") {
		t.Errorf("unexpected error for empty KDC host: %v", err)
	}

	if err := c.SessionSetupKerberosWithClient(nil, "cifs/host.corp.local"); err == nil {
		t.Error("expected an error for a nil Kerberos client")
	} else if !strings.Contains(err.Error(), "Kerberos client") {
		t.Errorf("unexpected error for nil Kerberos client: %v", err)
	}
}
