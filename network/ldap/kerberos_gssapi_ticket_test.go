package ldap

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	kerberos "github.com/TheManticoreProject/Manticore/network/kerberos/v5"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// forgeTGTFiles produces a ccache file and a .kirbi file holding a usable TGT, so
// the ticket-based bridge paths can be tested without a KDC.
func forgeTGTFiles(t *testing.T) (ccachePath, kirbiPath string) {
	t.Helper()

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}

	ft, err := kerberos.ForgeGolden(kerberos.ForgeOptions{
		Realm:     "CORP.LOCAL",
		Username:  "Administrator",
		DomainSID: "S-1-5-21-1-2-3",
		UserRID:   500,
		Key:       key,
		KeyEType:  messages.ETypeAES256CTSHMACSHA196,
	})
	if err != nil {
		t.Fatalf("ForgeGolden: %v", err)
	}

	kirbiBytes, err := ft.KirbiBytes()
	if err != nil {
		t.Fatalf("KirbiBytes: %v", err)
	}

	dir := t.TempDir()
	kirbiPath = filepath.Join(dir, "tgt.kirbi")
	if err := os.WriteFile(kirbiPath, kirbiBytes, 0o600); err != nil {
		t.Fatalf("write kirbi: %v", err)
	}

	// Round-trip the forged TGT through a client to serialize it as a ccache.
	kc := kerberos.NewClient("", "", "10.0.0.1")
	if err := kc.LoadTGTFromKirbiBytes(kirbiBytes); err != nil {
		t.Fatalf("LoadTGTFromKirbiBytes: %v", err)
	}
	cc, err := kc.ExportTGTCCache()
	if err != nil {
		t.Fatalf("ExportTGTCCache: %v", err)
	}
	ccachePath = filepath.Join(dir, "krb5cc")
	if err := cc.Save(ccachePath); err != nil {
		t.Fatalf("ccache.Save: %v", err)
	}
	return ccachePath, kirbiPath
}

// TestNewNativeGSSAPIClientLoadsTicket checks that a ccache or .kirbi credential
// makes the bridge load the TGT it carries, without an AS exchange. The KDC host is
// unreachable on purpose: if the bridge tried to acquire a TGT it would fail, so a
// nil error proves the ticket was used instead.
func TestNewNativeGSSAPIClientLoadsTicket(t *testing.T) {
	ccachePath, kirbiPath := forgeTGTFiles(t)

	// The forged TGT is for Administrator@CORP.LOCAL; the bridge must adopt that
	// principal from the ticket even though no username is passed in, so the AP-REQ
	// authenticator matches the ticket.
	check := func(t *testing.T, gss *nativeGSSAPIClient) {
		t.Helper()
		if gss == nil {
			t.Fatal("newNativeGSSAPIClient returned nil client")
		}
		if gss.user != "Administrator" {
			t.Errorf("adopted user = %q, want %q", gss.user, "Administrator")
		}
		if gss.realm != "CORP.LOCAL" {
			t.Errorf("adopted realm = %q, want %q", gss.realm, "CORP.LOCAL")
		}
	}

	t.Run("ccache", func(t *testing.T) {
		creds := &credentials.Credentials{}
		if err := creds.SetCCache(ccachePath); err != nil {
			t.Fatalf("SetCCache: %v", err)
		}
		gss, err := newNativeGSSAPIClient("10.0.0.1", "", creds)
		if err != nil {
			t.Fatalf("newNativeGSSAPIClient with ccache: %v", err)
		}
		check(t, gss)
	})

	t.Run("kirbi", func(t *testing.T) {
		creds := &credentials.Credentials{}
		if err := creds.SetKirbi(kirbiPath); err != nil {
			t.Fatalf("SetKirbi: %v", err)
		}
		gss, err := newNativeGSSAPIClient("10.0.0.1", "", creds)
		if err != nil {
			t.Fatalf("newNativeGSSAPIClient with kirbi: %v", err)
		}
		check(t, gss)
	})
}

// TestNewNativeGSSAPIClientRejectsBadTicket confirms an unreadable ccache surfaces
// as a ccache error, proving the branch routes to the ticket loader rather than a
// secret path.
func TestNewNativeGSSAPIClientRejectsBadTicket(t *testing.T) {
	path := filepath.Join(t.TempDir(), "krb5cc")
	if err := os.WriteFile(path, []byte("not a ccache"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	creds := &credentials.Credentials{}
	if err := creds.SetCCache(path); err != nil {
		t.Fatalf("SetCCache: %v", err)
	}
	_, err := newNativeGSSAPIClient("10.0.0.1", "CORP.LOCAL", creds)
	if err == nil {
		t.Fatal("newNativeGSSAPIClient with a garbage ccache error = nil, want an error")
	}
}
