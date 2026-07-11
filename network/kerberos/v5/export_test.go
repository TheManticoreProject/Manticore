package kerberos

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/credcache/ccache"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/credcache/kirbi"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// fakeTGTClient returns a client whose TGT fields are populated by hand, so the
// export path can be exercised without a live KDC.
func fakeTGTClient(t *testing.T) *KerberosClient {
	t.Helper()
	tkt := messages.Ticket{
		TktVno:  messages.KerberosV5,
		Realm:   "CORP.LOCAL",
		SName:   messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}},
		EncPart: messages.EncryptedData{EType: messages.ETypeAES256CTSHMACSHA196, Cipher: bytes.Repeat([]byte{0xAB}, 32)},
	}
	raw, err := tkt.Marshal()
	if err != nil {
		t.Fatalf("Ticket.Marshal: %v", err)
	}
	c := NewClient("alice", "corp.local", "10.0.0.1").WithPassword("x")
	c.tgtTicket = tkt
	c.tgtTicketRaw = raw
	c.sessionKey = bytes.Repeat([]byte{0x11}, 32)
	c.sessionEType = messages.ETypeAES256CTSHMACSHA196
	c.tgtEnc = messages.EncASRepPart{
		Flags:     messages.NewKerberosFlags(messages.TicketFlagForwardable, messages.TicketFlagRenewable, messages.TicketFlagInitial),
		AuthTime:  time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 7, 9, 20, 0, 0, 0, time.UTC),
		RenewTill: time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC),
		SRealm:    "CORP.LOCAL",
		SName:     messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}},
	}
	c.hasTGT = true
	return c
}

func TestExportTGTKirbi(t *testing.T) {
	c := fakeTGTClient(t)
	blob, err := c.ExportTGTKirbi()
	if err != nil {
		t.Fatalf("ExportTGTKirbi: %v", err)
	}
	cred, err := kirbi.Parse(blob)
	if err != nil {
		t.Fatalf("kirbi.Parse: %v", err)
	}
	if len(cred.Tickets) != 1 || cred.Tickets[0].Realm != "CORP.LOCAL" {
		t.Fatalf("exported kirbi tickets wrong: %+v", cred.Tickets)
	}
	infos, err := kirbi.TicketInfo(cred)
	if err != nil {
		t.Fatalf("TicketInfo: %v", err)
	}
	if infos[0].PName.NameString[0] != "alice" || !bytes.Equal(infos[0].Key.KeyValue, c.sessionKey) {
		t.Errorf("exported kirbi info wrong: %+v", infos[0])
	}
	if !infos[0].EndTime.Equal(c.tgtEnc.EndTime) {
		t.Errorf("endtime mismatch: %v vs %v", infos[0].EndTime, c.tgtEnc.EndTime)
	}
}

func TestExportTGTCCache(t *testing.T) {
	c := fakeTGTClient(t)
	cc, err := c.ExportTGTCCache()
	if err != nil {
		t.Fatalf("ExportTGTCCache: %v", err)
	}
	blob, err := cc.Marshal()
	if err != nil {
		t.Fatalf("ccache.Marshal: %v", err)
	}
	got, err := ccache.Unmarshal(blob)
	if err != nil {
		t.Fatalf("ccache.Unmarshal: %v", err)
	}
	if got.DefaultPrincipal.Realm != "CORP.LOCAL" || got.DefaultPrincipal.Components[0] != "alice" {
		t.Errorf("default principal wrong: %+v", got.DefaultPrincipal)
	}
	if len(got.Credentials) != 1 {
		t.Fatalf("credential count: %d", len(got.Credentials))
	}
	cr := got.Credentials[0]
	if cr.Server.Components[0] != "krbtgt" || cr.Key.EType != uint16(messages.ETypeAES256CTSHMACSHA196) {
		t.Errorf("credential fields wrong: %+v", cr)
	}
	if uint32(c.tgtEnc.EndTime.Unix()) != cr.EndTime {
		t.Errorf("endtime mismatch: %d vs %d", cr.EndTime, c.tgtEnc.EndTime.Unix())
	}
	if !bytes.Equal(cr.Ticket, c.tgtTicketRaw) {
		t.Errorf("ticket bytes not preserved")
	}
}

// cachedServiceTicketClient returns a client with a single cifs/host service
// ticket cached (as GetTGS would after a successful TGS-REP), so the
// service-ticket export path can be exercised without a live KDC.
func cachedServiceTicketClient(t *testing.T) (*KerberosClient, messages.EncTGSRepPart, []byte) {
	t.Helper()
	sname := messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"cifs", "host.corp.local"}}
	tkt := messages.Ticket{
		TktVno:  messages.KerberosV5,
		Realm:   "CORP.LOCAL",
		SName:   sname,
		EncPart: messages.EncryptedData{EType: messages.ETypeAES256CTSHMACSHA196, Cipher: bytes.Repeat([]byte{0xCD}, 32)},
	}
	raw, err := tkt.Marshal()
	if err != nil {
		t.Fatalf("Ticket.Marshal: %v", err)
	}
	enc := messages.EncTGSRepPart{
		Key:       messages.EncryptionKey{KeyType: messages.ETypeAES256CTSHMACSHA196, KeyValue: bytes.Repeat([]byte{0x22}, 32)},
		Flags:     messages.NewKerberosFlags(messages.TicketFlagForwardable, messages.TicketFlagRenewable),
		AuthTime:  time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC),
		StartTime: time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 7, 9, 20, 0, 0, 0, time.UTC),
		RenewTill: time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC),
		SRealm:    "CORP.LOCAL",
		SName:     sname,
	}
	c := NewClient("alice", "corp.local", "10.0.0.1").WithPassword("x")
	c.cacheServiceTicket(raw, enc)
	return c, enc, raw
}

func TestExportServiceTicketKirbi(t *testing.T) {
	c, enc, raw := cachedServiceTicketClient(t)
	blob, err := c.ExportServiceTicketKirbi("cifs/host.corp.local")
	if err != nil {
		t.Fatalf("ExportServiceTicketKirbi: %v", err)
	}
	st, err := LoadServiceTicketFromKirbiBytes(blob, "cifs/host.corp.local")
	if err != nil {
		t.Fatalf("LoadServiceTicketFromKirbiBytes: %v", err)
	}
	if !bytes.Equal(st.TicketRaw, raw) {
		t.Errorf("service ticket raw bytes not preserved")
	}
	if !bytes.Equal(st.SessionKey, enc.Key.KeyValue) {
		t.Errorf("session key mismatch")
	}
	if st.SessionEType != enc.Key.KeyType {
		t.Errorf("session etype: got %d, want %d", st.SessionEType, enc.Key.KeyType)
	}
	if st.SName.NameString[0] != "cifs" || st.SRealm != "CORP.LOCAL" {
		t.Errorf("sname/srealm wrong: %v %q", st.SName.NameString, st.SRealm)
	}
	if st.Client.NameString[0] != "alice" || st.CRealm != "CORP.LOCAL" {
		t.Errorf("client principal wrong: %v %q", st.Client.NameString, st.CRealm)
	}
}

func TestExportServiceTicketCCache(t *testing.T) {
	c, enc, raw := cachedServiceTicketClient(t)
	blob, err := c.ExportServiceTicketCCache("cifs/host.corp.local")
	if err != nil {
		t.Fatalf("ExportServiceTicketCCache: %v", err)
	}
	st, err := LoadServiceTicketFromCCacheBytes(blob, "cifs/host.corp.local")
	if err != nil {
		t.Fatalf("LoadServiceTicketFromCCacheBytes: %v", err)
	}
	if !bytes.Equal(st.TicketRaw, raw) {
		t.Errorf("service ticket raw bytes not preserved")
	}
	if !bytes.Equal(st.SessionKey, enc.Key.KeyValue) {
		t.Errorf("session key mismatch")
	}
	if st.SName.NameString[1] != "host.corp.local" || st.SRealm != "CORP.LOCAL" {
		t.Errorf("sname/srealm wrong: %v %q", st.SName.NameString, st.SRealm)
	}
	// The ccache credential carries the times/flags; confirm they survive.
	cc, err := ccache.Unmarshal(blob)
	if err != nil {
		t.Fatalf("ccache.Unmarshal: %v", err)
	}
	cr := cc.Credentials[0]
	if uint32(enc.EndTime.Unix()) != cr.EndTime || uint32(enc.RenewTill.Unix()) != cr.RenewTill {
		t.Errorf("times not preserved: end %d renew %d", cr.EndTime, cr.RenewTill)
	}
	if cr.TicketFlags != flagsToUint32(enc.Flags) {
		t.Errorf("flags not preserved: %08x vs %08x", cr.TicketFlags, flagsToUint32(enc.Flags))
	}
}

func TestExportServiceTicketToFile(t *testing.T) {
	c, _, raw := cachedServiceTicketClient(t)
	dir := t.TempDir()
	kpath := filepath.Join(dir, "svc.kirbi")
	cpath := filepath.Join(dir, "svc.ccache")
	if err := c.ExportServiceTicketKirbiToFile("cifs/host.corp.local", kpath); err != nil {
		t.Fatalf("ExportServiceTicketKirbiToFile: %v", err)
	}
	if err := c.ExportServiceTicketCCacheToFile("cifs/host.corp.local", cpath); err != nil {
		t.Fatalf("ExportServiceTicketCCacheToFile: %v", err)
	}
	kst, err := LoadServiceTicketFromKirbiFile(kpath, "cifs/host.corp.local")
	if err != nil {
		t.Fatalf("LoadServiceTicketFromKirbiFile: %v", err)
	}
	cst, err := LoadServiceTicketFromCCacheFile(cpath, "cifs/host.corp.local")
	if err != nil {
		t.Fatalf("LoadServiceTicketFromCCacheFile: %v", err)
	}
	if !bytes.Equal(kst.TicketRaw, raw) || !bytes.Equal(cst.TicketRaw, raw) {
		t.Errorf("ticket bytes not preserved through file round-trip")
	}
}

func TestExportServiceTicketRequiresCached(t *testing.T) {
	c := NewClient("alice", "corp.local", "10.0.0.1").WithPassword("x")
	if _, err := c.ExportServiceTicketKirbi("cifs/host.corp.local"); err == nil {
		t.Error("expected error exporting an uncached service ticket (kirbi)")
	}
	if _, err := c.ExportServiceTicketCCache("cifs/host.corp.local"); err == nil {
		t.Error("expected error exporting an uncached service ticket (ccache)")
	}
}

func TestExportRequiresTGT(t *testing.T) {
	c := NewClient("alice", "corp.local", "10.0.0.1").WithPassword("x")
	if _, err := c.ExportTGTKirbi(); err == nil {
		t.Error("expected error exporting without a TGT")
	}
	if _, err := c.ExportTGTCCache(); err == nil {
		t.Error("expected error exporting without a TGT")
	}
}
