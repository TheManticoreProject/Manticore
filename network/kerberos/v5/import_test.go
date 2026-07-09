package kerberos

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// assertUsableTGT checks that a freshly imported client carries a TGT state
// equivalent to the exporting client — the KDC-free half of pass-the-ticket.
func assertUsableTGT(t *testing.T, src, imp *KerberosClient) {
	t.Helper()
	if !imp.HasTGT() {
		t.Fatal("imported client has no TGT")
	}
	if imp.Realm() != src.Realm() {
		t.Errorf("realm: got %q, want %q", imp.Realm(), src.Realm())
	}
	if imp.Username() != src.Username() {
		t.Errorf("username: got %q, want %q", imp.Username(), src.Username())
	}
	if !bytes.Equal(imp.sessionKey, src.sessionKey) {
		t.Errorf("session key mismatch")
	}
	if imp.sessionEType != src.sessionEType {
		t.Errorf("session etype: got %d, want %d", imp.sessionEType, src.sessionEType)
	}
	if !bytes.Equal(imp.tgtTicketRaw, src.tgtTicketRaw) {
		t.Errorf("raw ticket bytes not preserved")
	}
	if imp.tgtTicket.Realm != src.tgtTicket.Realm {
		t.Errorf("parsed ticket realm: got %q, want %q", imp.tgtTicket.Realm, src.tgtTicket.Realm)
	}
	if !imp.tgtEnc.EndTime.Equal(src.tgtEnc.EndTime) {
		t.Errorf("endtime: got %v, want %v", imp.tgtEnc.EndTime, src.tgtEnc.EndTime)
	}
	// buildAPReq must succeed off the imported state alone — this is exactly what
	// GetTGS does before contacting the KDC.
	if _, err := imp.buildAPReq(); err != nil {
		t.Errorf("buildAPReq off imported TGT: %v", err)
	}
}

func TestImportTGTFromKirbiRoundTrip(t *testing.T) {
	src := fakeTGTClient(t)
	blob, err := src.ExportTGTKirbi()
	if err != nil {
		t.Fatalf("ExportTGTKirbi: %v", err)
	}
	// Fresh client with only a KDC host — identity comes from the ticket.
	imp := NewClient("", "", "10.0.0.1")
	if err := imp.LoadTGTFromKirbiBytes(blob); err != nil {
		t.Fatalf("LoadTGTFromKirbiBytes: %v", err)
	}
	assertUsableTGT(t, src, imp)
}

func TestImportTGTFromCCacheRoundTrip(t *testing.T) {
	src := fakeTGTClient(t)
	cc, err := src.ExportTGTCCache()
	if err != nil {
		t.Fatalf("ExportTGTCCache: %v", err)
	}
	blob, err := cc.Marshal()
	if err != nil {
		t.Fatalf("ccache.Marshal: %v", err)
	}
	imp := NewClient("", "", "10.0.0.1")
	if err := imp.LoadTGTFromCCacheBytes(blob); err != nil {
		t.Fatalf("LoadTGTFromCCacheBytes: %v", err)
	}
	assertUsableTGT(t, src, imp)
}

func TestImportTGTFromKirbiFile(t *testing.T) {
	src := fakeTGTClient(t)
	blob, err := src.ExportTGTKirbi()
	if err != nil {
		t.Fatalf("ExportTGTKirbi: %v", err)
	}
	path := filepath.Join(t.TempDir(), "tgt.kirbi")
	if err := os.WriteFile(path, blob, 0o600); err != nil {
		t.Fatalf("write kirbi: %v", err)
	}
	imp := NewClient("", "", "10.0.0.1")
	if err := imp.LoadTGTFromKirbiFile(path); err != nil {
		t.Fatalf("LoadTGTFromKirbiFile: %v", err)
	}
	assertUsableTGT(t, src, imp)
}

func TestImportTGTFromCCacheFileAndEnv(t *testing.T) {
	src := fakeTGTClient(t)
	cc, err := src.ExportTGTCCache()
	if err != nil {
		t.Fatalf("ExportTGTCCache: %v", err)
	}
	path := filepath.Join(t.TempDir(), "krb5cc")
	if err := cc.Save(path); err != nil {
		t.Fatalf("ccache.Save: %v", err)
	}

	imp := NewClient("", "", "10.0.0.1")
	if err := imp.LoadTGTFromCCacheFile(path); err != nil {
		t.Fatalf("LoadTGTFromCCacheFile: %v", err)
	}
	assertUsableTGT(t, src, imp)

	// KRB5CCNAME with a FILE: prefix.
	t.Setenv("KRB5CCNAME", "FILE:"+path)
	env := NewClient("", "", "10.0.0.1")
	if err := env.LoadTGTFromCCacheEnv(); err != nil {
		t.Fatalf("LoadTGTFromCCacheEnv: %v", err)
	}
	assertUsableTGT(t, src, env)
}

func TestImportServiceTicketFromKirbi(t *testing.T) {
	// Build a KRB-CRED holding a single service ticket (cifs/host).
	src := fakeServiceTicketClient(t)
	blob, err := src.ExportTGTKirbi() // fakeServiceTicketClient stores the ST as the "TGT" for export reuse
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	st, err := LoadServiceTicketFromKirbiBytes(blob, "cifs/host.corp.local")
	if err != nil {
		t.Fatalf("LoadServiceTicketFromKirbiBytes: %v", err)
	}
	if !bytes.Equal(st.TicketRaw, src.tgtTicketRaw) {
		t.Errorf("service ticket raw bytes not preserved")
	}
	if !bytes.Equal(st.SessionKey, src.sessionKey) {
		t.Errorf("service ticket session key mismatch")
	}
	if st.SName.NameString[0] != "cifs" {
		t.Errorf("service name: got %v", st.SName.NameString)
	}
	// A non-matching SPN must not resolve.
	if _, err := LoadServiceTicketFromKirbiBytes(blob, "host/other"); err == nil {
		t.Error("expected no match for host/other")
	}
	// Empty SPN takes the first ticket.
	if _, err := LoadServiceTicketFromKirbiBytes(blob, ""); err != nil {
		t.Errorf("empty SPN should take the first ticket: %v", err)
	}
}

func TestImportServiceTicketFromCCache(t *testing.T) {
	src := fakeServiceTicketClient(t)
	cc, err := src.ExportTGTCCache()
	if err != nil {
		t.Fatalf("export ccache: %v", err)
	}
	blob, err := cc.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	st, err := LoadServiceTicketFromCCacheBytes(blob, "cifs/host.corp.local")
	if err != nil {
		t.Fatalf("LoadServiceTicketFromCCacheBytes: %v", err)
	}
	if !bytes.Equal(st.TicketRaw, src.tgtTicketRaw) {
		t.Errorf("service ticket raw bytes not preserved")
	}
	if st.SRealm != "CORP.LOCAL" {
		t.Errorf("srealm: got %q", st.SRealm)
	}
}

func TestImportRejectsEncryptedKirbi(t *testing.T) {
	src := fakeTGTClient(t)
	cred, err := src.tgtCredInfoKirbi()
	if err != nil {
		t.Fatalf("build kirbi: %v", err)
	}
	// Flip the enc-part to an encrypted etype: the session key is unreadable.
	cred.EncPart.EType = messages.ETypeAES256CTSHMACSHA196
	blob, err := cred.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	imp := NewClient("", "", "10.0.0.1")
	if err := imp.LoadTGTFromKirbiBytes(blob); err == nil {
		t.Error("expected error loading an encrypted kirbi enc-part")
	}
}

// fakeServiceTicketClient returns a client whose exported ticket is a cifs/host
// service ticket (reusing the export path to produce a valid container).
func fakeServiceTicketClient(t *testing.T) *KerberosClient {
	t.Helper()
	c := fakeTGTClient(t)
	sname := messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"cifs", "host.corp.local"}}
	tkt := c.tgtTicket
	tkt.SName = sname
	raw, err := tkt.Marshal()
	if err != nil {
		t.Fatalf("marshal service ticket: %v", err)
	}
	c.tgtTicket = tkt
	c.tgtTicketRaw = raw
	c.tgtEnc.SName = sname
	c.tgtEnc.SRealm = "CORP.LOCAL"
	return c
}

// tgtCredInfoKirbi rebuilds the .kirbi KRB-CRED for the current TGT so a test can
// mutate it before marshaling.
func (c *KerberosClient) tgtCredInfoKirbi() (*messages.KRBCred, error) {
	enc := messages.EncKrbCredPart{TicketInfo: []messages.KrbCredInfo{c.tgtCredInfo()}}
	encBytes, err := enc.Marshal()
	if err != nil {
		return nil, err
	}
	return &messages.KRBCred{
		PVNO:       messages.KerberosV5,
		MsgType:    messages.MsgTypeKRBCred,
		TicketsRaw: [][]byte{c.tgtTicketRaw},
		EncPart:    messages.EncryptedData{EType: 0, Cipher: encBytes},
	}, nil
}
