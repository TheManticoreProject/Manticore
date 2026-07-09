package messages

import (
	"bytes"
	"encoding/asn1"
	"testing"
	"time"
)

func TestAPRepRoundtrip(t *testing.T) {
	subkey := &EncryptionKey{KeyType: ETypeAES256CTSHMACSHA196, KeyValue: bytes.Repeat([]byte{0x5a}, 32)}
	orig := &EncAPRepPart{
		CTime:     time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC),
		CUSec:     123456,
		SubKey:    subkey,
		SeqNumber: 0x11223344,
	}
	encPartBytes, err := orig.Marshal()
	if err != nil {
		t.Fatalf("EncAPRepPart.Marshal: %v", err)
	}
	var decEnc EncAPRepPart
	if _, err := decEnc.Unmarshal(encPartBytes); err != nil {
		t.Fatalf("EncAPRepPart.Unmarshal: %v", err)
	}
	if !decEnc.CTime.Equal(orig.CTime) || decEnc.CUSec != orig.CUSec || decEnc.SeqNumber != orig.SeqNumber {
		t.Errorf("EncAPRepPart mismatch: %+v vs %+v", decEnc, orig)
	}
	if decEnc.SubKey == nil || decEnc.SubKey.KeyType != subkey.KeyType || !bytes.Equal(decEnc.SubKey.KeyValue, subkey.KeyValue) {
		t.Errorf("EncAPRepPart subkey mismatch: %+v", decEnc.SubKey)
	}

	ap := &APRep{EncPart: EncryptedData{EType: ETypeAES256CTSHMACSHA196, Cipher: []byte{1, 2, 3, 4}}}
	wire, err := ap.Marshal()
	if err != nil {
		t.Fatalf("APRep.Marshal: %v", err)
	}
	var decAP APRep
	if _, err := decAP.Unmarshal(wire); err != nil {
		t.Fatalf("APRep.Unmarshal: %v", err)
	}
	if decAP.PVNO != KerberosV5 || decAP.MsgType != MsgTypeAPRep {
		t.Errorf("APRep header: pvno=%d msgtype=%d", decAP.PVNO, decAP.MsgType)
	}
	if !bytes.Equal(decAP.EncPart.Cipher, ap.EncPart.Cipher) {
		t.Errorf("APRep enc-part cipher mismatch")
	}
}

func TestKRBCredRoundtrip(t *testing.T) {
	tkt := Ticket{
		TktVno: KerberosV5,
		Realm:  "CORP.LOCAL",
		SName:  PrincipalName{NameType: NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}},
		EncPart: EncryptedData{
			EType:  ETypeAES256CTSHMACSHA196,
			Cipher: bytes.Repeat([]byte{0xAB}, 32),
		},
	}
	tktRaw, err := tkt.Marshal()
	if err != nil {
		t.Fatalf("Ticket.Marshal: %v", err)
	}

	enc := EncKrbCredPart{
		TicketInfo: []KrbCredInfo{{
			Key:       EncryptionKey{KeyType: ETypeAES256CTSHMACSHA196, KeyValue: bytes.Repeat([]byte{0x11}, 32)},
			PRealm:    "CORP.LOCAL",
			PName:     PrincipalName{NameType: NameTypePrincipal, NameString: []string{"alice"}},
			Flags:     NewKerberosFlags(TicketFlagForwardable, TicketFlagRenewable, TicketFlagInitial),
			AuthTime:  time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC),
			StartTime: time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 7, 9, 20, 0, 0, 0, time.UTC),
			RenewTill: time.Date(2026, 7, 16, 10, 0, 0, 0, time.UTC),
			SRealm:    "CORP.LOCAL",
			SName:     PrincipalName{NameType: NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}},
		}},
	}
	encBytes, err := enc.Marshal()
	if err != nil {
		t.Fatalf("EncKrbCredPart.Marshal: %v", err)
	}
	var decEnc EncKrbCredPart
	if _, err := decEnc.Unmarshal(encBytes); err != nil {
		t.Fatalf("EncKrbCredPart.Unmarshal: %v", err)
	}
	if len(decEnc.TicketInfo) != 1 {
		t.Fatalf("ticket-info count: got %d, want 1", len(decEnc.TicketInfo))
	}
	got := decEnc.TicketInfo[0]
	want := enc.TicketInfo[0]
	if got.PRealm != want.PRealm || got.SRealm != want.SRealm {
		t.Errorf("realm mismatch: prealm=%q srealm=%q", got.PRealm, got.SRealm)
	}
	if len(got.PName.NameString) != 1 || got.PName.NameString[0] != "alice" {
		t.Errorf("pname mismatch: %+v", got.PName)
	}
	if len(got.SName.NameString) != 2 || got.SName.NameString[0] != "krbtgt" {
		t.Errorf("sname mismatch: %+v", got.SName)
	}
	if !got.EndTime.Equal(want.EndTime) || !got.RenewTill.Equal(want.RenewTill) {
		t.Errorf("time mismatch: end=%v renew=%v", got.EndTime, got.RenewTill)
	}
	if !bytes.Equal(got.Key.KeyValue, want.Key.KeyValue) {
		t.Errorf("session key mismatch")
	}

	cred := &KRBCred{
		Tickets:    []Ticket{tkt},
		TicketsRaw: [][]byte{tktRaw},
		EncPart:    EncryptedData{EType: 0, Cipher: encBytes},
	}
	wire, err := cred.Marshal()
	if err != nil {
		t.Fatalf("KRBCred.Marshal: %v", err)
	}
	var decCred KRBCred
	if _, err := decCred.Unmarshal(wire); err != nil {
		t.Fatalf("KRBCred.Unmarshal: %v", err)
	}
	if decCred.PVNO != KerberosV5 || decCred.MsgType != MsgTypeKRBCred {
		t.Errorf("KRBCred header: pvno=%d msgtype=%d", decCred.PVNO, decCred.MsgType)
	}
	if len(decCred.Tickets) != 1 || decCred.Tickets[0].Realm != "CORP.LOCAL" {
		t.Fatalf("KRBCred tickets mismatch: %+v", decCred.Tickets)
	}
	// The embedded EncKrbCredPart must survive the full KRB-CRED round-trip.
	var reEnc EncKrbCredPart
	if _, err := reEnc.Unmarshal(decCred.EncPart.Cipher); err != nil {
		t.Fatalf("re-parse EncKrbCredPart: %v", err)
	}
	if len(reEnc.TicketInfo) != 1 || reEnc.TicketInfo[0].PName.NameString[0] != "alice" {
		t.Errorf("embedded enc-part mismatch after KRBCred round-trip")
	}
}

// TestNewKerberosFlags verifies 32-bit MSB-first flag encoding.
func TestNewKerberosFlags(t *testing.T) {
	f := NewKerberosFlags(TicketFlagForwardable) // bit 1 → 0x40 in byte 0
	if f.BitLength != 32 || len(f.Bytes) != 4 {
		t.Fatalf("flags not 32-bit: bitlen=%d bytes=%d", f.BitLength, len(f.Bytes))
	}
	if f.Bytes[0] != 0x40 || f.Bytes[1] != 0 || f.Bytes[2] != 0 || f.Bytes[3] != 0 {
		t.Errorf("bit 1 encoding wrong: % X", f.Bytes)
	}
	// The 32-bit width must survive a DER round-trip (no trailing-zero truncation).
	b, err := asn1.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	var back asn1.BitString
	if _, err := asn1.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.BitLength != 32 {
		t.Errorf("DER round-trip lost width: bitlen=%d", back.BitLength)
	}
}

// TestNormalizeTimeUTC confirms a non-UTC time marshals with a "Z" suffix.
func TestNormalizeTimeUTC(t *testing.T) {
	loc := time.FixedZone("EST", -5*3600)
	nonUTC := time.Date(2026, 7, 9, 8, 0, 0, 0, loc)
	e := &EncAPRepPart{CTime: nonUTC, CUSec: 1}
	b, err := e.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	// The GeneralizedTime inside must end in 'Z' (0x5A), never a numeric offset.
	if !bytes.Contains(b, []byte("Z")) {
		t.Errorf("marshaled time is not UTC (no 'Z'): % X", b)
	}
	if bytes.Contains(b, []byte("-0500")) || bytes.Contains(b, []byte("+")) {
		t.Errorf("marshaled time carries a numeric offset: % X", b)
	}
}
