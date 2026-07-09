package kirbi

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

func sampleTicketAndInfo(t *testing.T) ([]byte, messages.KrbCredInfo) {
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
	info := messages.KrbCredInfo{
		Key:       messages.EncryptionKey{KeyType: messages.ETypeAES256CTSHMACSHA196, KeyValue: bytes.Repeat([]byte{0x11}, 32)},
		PRealm:    "CORP.LOCAL",
		PName:     messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{"alice"}},
		Flags:     messages.NewKerberosFlags(messages.TicketFlagForwardable, messages.TicketFlagRenewable, messages.TicketFlagInitial),
		AuthTime:  time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 7, 9, 20, 0, 0, 0, time.UTC),
		RenewTill: time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC),
		SRealm:    "CORP.LOCAL",
		SName:     messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}},
	}
	return raw, info
}

func TestKirbiNewRoundtrip(t *testing.T) {
	raw, info := sampleTicketAndInfo(t)

	cred, err := New(raw, info)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if cred.EncPart.EType != 0 {
		t.Errorf("enc-part should be unencrypted (etype 0), got %d", cred.EncPart.EType)
	}

	blob, err := Bytes(cred)
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	// Sanity: a .kirbi is an APPLICATION[22] (KRB-CRED) — first byte 0x76.
	if blob[0] != 0x76 {
		t.Errorf("kirbi should start with APPLICATION[22] tag 0x76, got 0x%02x", blob[0])
	}

	parsed, err := Parse(blob)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(parsed.Tickets) != 1 || parsed.Tickets[0].Realm != "CORP.LOCAL" {
		t.Fatalf("parsed tickets wrong: %+v", parsed.Tickets)
	}

	infos, err := TicketInfo(parsed)
	if err != nil {
		t.Fatalf("TicketInfo: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("ticket-info count: got %d, want 1", len(infos))
	}
	got := infos[0]
	if got.PName.NameString[0] != "alice" || got.SRealm != "CORP.LOCAL" {
		t.Errorf("ticket-info fields wrong: %+v", got)
	}
	if !got.EndTime.Equal(info.EndTime) || !got.RenewTill.Equal(info.RenewTill) {
		t.Errorf("time round-trip: end=%v renew=%v", got.EndTime, got.RenewTill)
	}
	if !bytes.Equal(got.Key.KeyValue, info.Key.KeyValue) {
		t.Error("session key mismatch")
	}
}

func TestKirbiSaveLoad(t *testing.T) {
	raw, info := sampleTicketAndInfo(t)
	cred, err := New(raw, info)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "ticket.kirbi")
	if err := Save(path, cred); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Tickets) != 1 {
		t.Fatalf("loaded tickets: %d", len(loaded.Tickets))
	}
	// The raw ticket bytes must survive verbatim.
	if !bytes.Equal(loaded.TicketsRaw[0], raw) {
		t.Error("raw ticket bytes not preserved through save/load")
	}
}

func TestTicketInfoRejectsEncrypted(t *testing.T) {
	cred := &messages.KRBCred{
		PVNO:    messages.KerberosV5,
		MsgType: messages.MsgTypeKRBCred,
		EncPart: messages.EncryptedData{EType: messages.ETypeAES256CTSHMACSHA196, Cipher: []byte{1, 2, 3}},
	}
	if _, err := TicketInfo(cred); err == nil {
		t.Error("expected error reading ticket-info from an encrypted enc-part")
	}
}
