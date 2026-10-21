package client

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

func passwordCreds(t *testing.T, domain string) *credentials.Credentials {
	t.Helper()
	creds, err := credentials.NewCredentials(domain, "alice", "Password1!", "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}
	return creds
}

// TestLoginKerberosSPN checks the target principal handed to the backend: the
// one given, or "cifs/<host>" built from the name Dial was given.
func TestLoginKerberosSPN(t *testing.T) {
	rb := &recordingBackend{}
	c := &Client{backend: rb, host: "fs01.corp.local"}

	if err := c.LoginKerberos(passwordCreds(t, "corp.local"), KerberosOptions{KDCHost: "dc01.corp.local"}); err != nil {
		t.Fatalf("LoginKerberos: %v", err)
	}
	if err := c.LoginKerberos(passwordCreds(t, "corp.local"), KerberosOptions{KDCHost: "dc01.corp.local", SPN: "cifs/alias.corp.local"}); err != nil {
		t.Fatalf("LoginKerberos: %v", err)
	}
	want := []string{"cifs/fs01.corp.local", "cifs/alias.corp.local"}
	if len(rb.args) != 2 || rb.args[0] != want[0] || rb.args[1] != want[1] {
		t.Errorf("SPNs = %v, want %v", rb.args, want)
	}
}

// TestNewKerberosClientIdentity checks that the realm is the upper-cased domain,
// as the KDC spells it, and that the username and KDC are carried over.
func TestNewKerberosClientIdentity(t *testing.T) {
	kc, err := newKerberosClient(passwordCreds(t, "corp.local"), "dc01.corp.local:88")
	if err != nil {
		t.Fatalf("newKerberosClient: %v", err)
	}
	if kc.Realm() != "CORP.LOCAL" || kc.Username() != "alice" || kc.KDCHost() != "dc01.corp.local:88" {
		t.Errorf("client = %s@%s via %s, want alice@CORP.LOCAL via dc01.corp.local:88", kc.Username(), kc.Realm(), kc.KDCHost())
	}
}

func TestNewKerberosClientSecrets(t *testing.T) {
	t.Setenv("KRB5CCNAME", "")

	// No secret and no KRB5CCNAME.
	creds, _ := credentials.NewCredentials("corp.local", "alice", "", "")
	if _, err := newKerberosClient(creds, ""); err == nil || !strings.Contains(err.Error(), "no secret") {
		t.Errorf("no secret: err = %v, want a 'no secret' error", err)
	}

	// KRB5CCNAME naming a missing cache.
	t.Setenv("KRB5CCNAME", "FILE:"+filepath.Join(t.TempDir(), "missing.ccache"))
	if _, err := newKerberosClient(creds, ""); err == nil || !strings.Contains(err.Error(), "KRB5CCNAME") {
		t.Errorf("missing KRB5CCNAME cache: err = %v, want a KRB5CCNAME error", err)
	}

	// A malformed AES key is reported, not silently ignored.
	creds, _ = credentials.NewCredentials("corp.local", "alice", "", "")
	creds.AESKey = "not-hex"
	if _, err := newKerberosClient(creds, ""); err == nil || !strings.Contains(err.Error(), "AES key") {
		t.Errorf("bad AES key: err = %v, want an AES key error", err)
	}

	// An NT hash selects overpass-the-hash.
	creds, _ = credentials.NewCredentials("corp.local", "alice", "", ":8846f7eaee8fb117ad06bdd830b7586c")
	if _, err := newKerberosClient(creds, ""); err != nil {
		t.Errorf("NT hash: %v", err)
	}
}

func TestSMB1LoginKerberosUnsupported(t *testing.T) {
	if err := (&smb1Backend{}).LoginKerberos(nil, "cifs/host"); err == nil {
		t.Error("SMB1 backend accepted a Kerberos login")
	}
}
