package kerberos

import (
	"bytes"
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// RFC 6113 §5.4.3: the client must verify the KrbFastFinished ticket-checksum — a
// keyed checksum over the issued ticket, computed with the armor key under key usage
// 53 — and take the client identity from the finished structure rather than the
// unprotected outer reply. These tests cover verifyFASTFinished, the shared check
// both the FAST AS and TGS reply processors run.

// fastFinishedFixture returns an armor key, the ticket bytes it authenticates, and a
// KrbFastFinished whose ticket-checksum is valid for them.
func fastFinishedFixture(t *testing.T) (armorKey, ticketRaw []byte, armorEType int, finished *messages.KrbFastFinished) {
	t.Helper()
	armorEType = messages.ETypeAES256CTSHMACSHA196
	armorKey = bytes.Repeat([]byte{0x42}, kerbcrypto.KeyLen(armorEType))
	ticketRaw = bytes.Repeat([]byte{0xA5}, 96) // stand-in for the APPLICATION 1 ticket

	required, ok := kerbcrypto.ChecksumTypeForEType(armorEType)
	if !ok {
		t.Fatalf("no checksum type for etype %d", armorEType)
	}
	cksum, err := kerbcrypto.GetChecksum(required, armorKey, kerbcrypto.KeyUsageFASTFinished, ticketRaw)
	if err != nil {
		t.Fatalf("GetChecksum: %v", err)
	}
	finished = &messages.KrbFastFinished{
		CRealm:         "CORP.LOCAL",
		CName:          messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{"alice"}},
		TicketChecksum: messages.Checksum{CKSumType: required, Checksum: cksum},
	}
	return armorKey, ticketRaw, armorEType, finished
}

func TestVerifyFASTFinished_ValidReturnsAuthenticatedIdentity(t *testing.T) {
	c := NewClient("alice", "CORP.LOCAL", "10.0.0.1")
	armorKey, ticketRaw, armorEType, finished := fastFinishedFixture(t)

	crealm, cname, err := c.verifyFASTFinished(finished, ticketRaw, armorKey, armorEType, "FAST AS-REP")
	if err != nil {
		t.Fatalf("verifyFASTFinished on a valid finished: %v", err)
	}
	// The identity must come from the (armor-authenticated) finished, not elsewhere.
	if crealm != "CORP.LOCAL" {
		t.Errorf("crealm = %q, want CORP.LOCAL", crealm)
	}
	if len(cname.NameString) != 1 || cname.NameString[0] != "alice" {
		t.Errorf("cname = %+v, want alice", cname)
	}
}

func TestVerifyFASTFinished_RejectsTamperedTicket(t *testing.T) {
	c := NewClient("alice", "CORP.LOCAL", "10.0.0.1")
	armorKey, ticketRaw, armorEType, finished := fastFinishedFixture(t)

	// An on-path attacker swaps the issued ticket for a forged one: the checksum no
	// longer covers the bytes presented, so it must be rejected.
	tampered := append([]byte(nil), ticketRaw...)
	tampered[0] ^= 0xFF
	if _, _, err := c.verifyFASTFinished(finished, tampered, armorKey, armorEType, "FAST AS-REP"); err == nil {
		t.Fatal("verifyFASTFinished accepted a ticket the checksum does not cover")
	}
}

func TestVerifyFASTFinished_RejectsTamperedChecksum(t *testing.T) {
	c := NewClient("alice", "CORP.LOCAL", "10.0.0.1")
	armorKey, ticketRaw, armorEType, finished := fastFinishedFixture(t)

	finished.TicketChecksum.Checksum[0] ^= 0xFF
	if _, _, err := c.verifyFASTFinished(finished, ticketRaw, armorKey, armorEType, "FAST AS-REP"); err == nil {
		t.Fatal("verifyFASTFinished accepted a corrupted ticket-checksum")
	}
}

func TestVerifyFASTFinished_RejectsWrongArmorKey(t *testing.T) {
	c := NewClient("alice", "CORP.LOCAL", "10.0.0.1")
	_, ticketRaw, armorEType, finished := fastFinishedFixture(t)

	// A different armor key (an attacker without the armor TGT) cannot have produced
	// the checksum.
	wrongKey := bytes.Repeat([]byte{0x11}, kerbcrypto.KeyLen(armorEType))
	if _, _, err := c.verifyFASTFinished(finished, ticketRaw, wrongKey, armorEType, "FAST AS-REP"); err == nil {
		t.Fatal("verifyFASTFinished accepted a checksum under the wrong armor key")
	}
}

func TestVerifyFASTFinished_RejectsUnkeyedChecksumType(t *testing.T) {
	c := NewClient("alice", "CORP.LOCAL", "10.0.0.1")
	armorKey, ticketRaw, armorEType, _ := fastFinishedFixture(t)

	// A finished that advertises an unkeyed checksum type (CRC32 — forgeable without
	// the armor key) must be rejected on the type pin, before the checksum value is
	// even evaluated, because that type is not the required keyed type for the armor
	// enctype. Any bytes suffice since the type is wrong.
	downgraded := &messages.KrbFastFinished{
		CRealm:         "CORP.LOCAL",
		CName:          messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{"alice"}},
		TicketChecksum: messages.Checksum{CKSumType: iana.CksumTypeCRC32, Checksum: bytes.Repeat([]byte{0x00}, 4)},
	}
	if _, _, err := c.verifyFASTFinished(downgraded, ticketRaw, armorKey, armorEType, "FAST AS-REP"); err == nil {
		t.Fatal("verifyFASTFinished accepted a non-required (unkeyed) checksum type")
	}
}

func TestVerifyFASTFinished_RejectsMissingFinished(t *testing.T) {
	c := NewClient("alice", "CORP.LOCAL", "10.0.0.1")
	armorKey, ticketRaw, armorEType, _ := fastFinishedFixture(t)

	if _, _, err := c.verifyFASTFinished(nil, ticketRaw, armorKey, armorEType, "FAST AS-REP"); err == nil {
		t.Fatal("verifyFASTFinished accepted a reply with no KrbFastFinished")
	}
}
