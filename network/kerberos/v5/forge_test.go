package kerberos

import (
	"bytes"
	"encoding/asn1"
	"testing"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/pac"
)

const (
	testRealm     = "CORP.LOCAL"
	testDomainSID = "S-1-5-21-1111111111-2222222222-3333333333"
)

// extractPAC decrypts a forged ticket enc-part with key, parses it, and returns
// the embedded PAC bytes (AD-IF-RELEVANT → AD-WIN2K-PAC).
func extractPAC(t *testing.T, tk messages.Ticket, key []byte, etype int) (*messages.EncTicketPart, []byte) {
	t.Helper()
	plain, err := kerbcrypto.Decrypt(etype, key, kerbcrypto.KeyUsageKDCRepTicket, tk.EncPart.Cipher)
	if err != nil {
		t.Fatalf("decrypt enc-part: %v", err)
	}
	var enc messages.EncTicketPart
	if _, err := enc.Unmarshal(plain); err != nil {
		t.Fatalf("unmarshal EncTicketPart: %v", err)
	}
	if len(enc.AuthorizationData) != 1 || enc.AuthorizationData[0].ADType != 1 {
		t.Fatalf("expected one AD-IF-RELEVANT element, got %+v", enc.AuthorizationData)
	}
	var inner []messages.AuthorizationData
	if _, err := asn1.Unmarshal(enc.AuthorizationData[0].ADData, &inner); err != nil {
		t.Fatalf("unmarshal AD-IF-RELEVANT contents: %v", err)
	}
	for _, e := range inner {
		if e.ADType == 128 {
			return &enc, e.ADData
		}
	}
	t.Fatal("no AD-WIN2K-PAC element")
	return nil, nil
}

func TestForgeGoldenStructure(t *testing.T) {
	key := bytes.Repeat([]byte{0xAB}, 16) // krbtgt RC4 NT hash

	ft, err := ForgeGolden(ForgeOptions{
		Realm:           testRealm,
		Username:        "Administrator",
		DomainSID:       testDomainSID,
		UserRID:         500,
		LogonDomainName: "CORP",
		LogonServer:     "DC01",
		Key:             key,
		KeyEType:        23,
		StartTime:       time.Unix(1700000000, 0),
	})
	if err != nil {
		t.Fatalf("ForgeGolden: %v", err)
	}

	// The service name must be krbtgt/REALM.
	if ft.Ticket.SName.NameType != messages.NameTypeSRVInst ||
		len(ft.Ticket.SName.NameString) != 2 ||
		ft.Ticket.SName.NameString[0] != "krbtgt" ||
		ft.Ticket.SName.NameString[1] != testRealm {
		t.Errorf("golden SName = %+v, want krbtgt/%s", ft.Ticket.SName, testRealm)
	}
	if ft.Ticket.EncPart.EType != 23 {
		t.Errorf("enc-part etype = %d, want 23", ft.Ticket.EncPart.EType)
	}

	enc, pacBytes := extractPAC(t, ft.Ticket, key, 23)
	if enc.CRealm != testRealm {
		t.Errorf("CRealm = %q, want %q", enc.CRealm, testRealm)
	}
	if len(enc.CName.NameString) != 1 || enc.CName.NameString[0] != "Administrator" {
		t.Errorf("CName = %+v, want Administrator", enc.CName)
	}
	if !enc.EndTime.After(enc.StartTime) {
		t.Errorf("EndTime %v not after StartTime %v", enc.EndTime, enc.StartTime)
	}

	p, err := pac.Parse(pacBytes)
	if err != nil {
		t.Fatalf("parse embedded PAC: %v", err)
	}
	if err := p.VerifyServerSignature(key); err != nil {
		t.Errorf("PAC server signature: %v", err)
	}
	if err := p.VerifyKDCSignature(key); err != nil {
		t.Errorf("PAC KDC signature: %v", err)
	}

	// The forged golden ticket must round-trip through the import path.
	kirbi, err := ft.KirbiBytes()
	if err != nil {
		t.Fatalf("KirbiBytes: %v", err)
	}
	c := NewClient("", "", "")
	if err := c.LoadTGTFromKirbiBytes(kirbi); err != nil {
		t.Fatalf("LoadTGTFromKirbiBytes: %v", err)
	}
	if c.Username() != "Administrator" || c.Realm() != testRealm {
		t.Errorf("imported identity = %q@%q, want Administrator@%s", c.Username(), c.Realm(), testRealm)
	}
}

func TestForgeSilverStructure(t *testing.T) {
	key := bytes.Repeat([]byte{0xCD}, 16) // service account RC4 NT hash
	spn := "cifs/host.corp.local"

	ft, err := ForgeSilver(ForgeOptions{
		Realm:           testRealm,
		Username:        "Administrator",
		DomainSID:       testDomainSID,
		UserRID:         500,
		LogonDomainName: "CORP",
		Key:             key,
		KeyEType:        23,
	}, spn)
	if err != nil {
		t.Fatalf("ForgeSilver: %v", err)
	}

	if len(ft.Ticket.SName.NameString) != 2 || ft.Ticket.SName.NameString[0] != "cifs" {
		t.Errorf("silver SName = %+v, want cifs/host.corp.local", ft.Ticket.SName)
	}

	enc, pacBytes := extractPAC(t, ft.Ticket, key, 23)
	// A silver ticket is a service ticket: it must NOT be flagged initial.
	if bitSet(enc.Flags, messages.TicketFlagInitial) {
		t.Error("silver ticket should not carry the initial flag")
	}
	p, err := pac.Parse(pacBytes)
	if err != nil {
		t.Fatalf("parse embedded PAC: %v", err)
	}
	if err := p.VerifyServerSignature(key); err != nil {
		t.Errorf("PAC server signature: %v", err)
	}
}

func TestForgeGoldenIsInitial(t *testing.T) {
	key := bytes.Repeat([]byte{0x11}, 16)
	ft, err := ForgeGolden(ForgeOptions{
		Realm: testRealm, Username: "u", DomainSID: testDomainSID, Key: key, KeyEType: 23,
	})
	if err != nil {
		t.Fatalf("ForgeGolden: %v", err)
	}
	enc, _ := extractPAC(t, ft.Ticket, key, 23)
	if !bitSet(enc.Flags, messages.TicketFlagInitial) {
		t.Error("golden TGT must carry the initial flag")
	}
	if !bitSet(enc.Flags, messages.TicketFlagRenewable) {
		t.Error("golden TGT should be renewable")
	}
}

func TestForgeValidation(t *testing.T) {
	_, err := ForgeGolden(ForgeOptions{Realm: testRealm, Username: "u", Key: []byte{1}, KeyEType: 23})
	if err == nil {
		t.Error("expected error when DomainSID is missing")
	}
	_, err = ForgeGolden(ForgeOptions{Realm: testRealm, Username: "u", DomainSID: testDomainSID, KeyEType: 23})
	if err == nil {
		t.Error("expected error when Key is missing")
	}
	_, err = ForgeGolden(ForgeOptions{Realm: testRealm, Username: "u", DomainSID: testDomainSID, Key: []byte{1}, KeyEType: 1})
	if err == nil {
		t.Error("expected error for unsupported signing-key etype")
	}
}

// bitSet reports whether the given ticket-flag bit is set in an MSB-first
// KerberosFlags BIT STRING.
func bitSet(flags asn1.BitString, bit int) bool {
	return flags.At(bit) == 1
}
