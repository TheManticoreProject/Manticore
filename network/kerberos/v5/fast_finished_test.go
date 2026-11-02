package kerberos

import (
	"bytes"
	"strings"
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// fastFinishedFor builds a KrbFastResponse whose KrbFastFinished names
// cname@crealm and checksums ticketRaw under armorKey (key usage 53).
func fastFinishedFor(t *testing.T, ticketRaw, armorKey []byte, crealm, cname string) *messages.KrbFastResponse {
	t.Helper()
	sum, err := kerbcrypto.GetChecksum(iana.CksumTypeHMACSHA196AES256, armorKey, kerbcrypto.KeyUsageFASTFinished, ticketRaw)
	if err != nil {
		t.Fatal(err)
	}
	return &messages.KrbFastResponse{Finished: &messages.KrbFastFinished{
		CRealm:         crealm,
		CName:          messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{cname}},
		TicketChecksum: messages.Checksum{CKSumType: iana.CksumTypeHMACSHA196AES256, Checksum: sum},
	}}
}

// TestVerifyFASTFinished checks the RFC 6113 section 5.4.3 rules: the
// ticket-checksum must match the ticket as received, a reply without
// KrbFastFinished is refused, and the client identity comes from finished.
func TestVerifyFASTFinished(t *testing.T) {
	armorKey := bytes.Repeat([]byte{0x42}, 32)
	ticketRaw := []byte{0x61, 0x03, 0x02, 0x01, 0x05}

	resp := fastFinishedFor(t, ticketRaw, armorKey, "CORP.LOCAL", "alice")
	realm, name, err := verifyFASTFinished("FAST AS-REP", resp, ticketRaw, armorKey)
	if err != nil {
		t.Fatalf("valid finished rejected: %v", err)
	}
	if realm != "CORP.LOCAL" || strings.Join(name.NameString, "/") != "alice" {
		t.Errorf("identity = %s@%s, want alice@CORP.LOCAL from finished", strings.Join(name.NameString, "/"), realm)
	}

	tampered := append([]byte(nil), ticketRaw...)
	tampered[4] ^= 0x01
	if _, _, err := verifyFASTFinished("FAST AS-REP", resp, tampered, armorKey); err == nil {
		t.Error("a ticket that does not match ticket-checksum was accepted")
	}

	if _, _, err := verifyFASTFinished("FAST AS-REP", resp, ticketRaw, bytes.Repeat([]byte{0x43}, 32)); err == nil {
		t.Error("a ticket-checksum under another key was accepted")
	}

	if _, _, err := verifyFASTFinished("FAST TGS-REP", &messages.KrbFastResponse{}, ticketRaw, armorKey); err == nil {
		t.Error("a reply without KrbFastFinished was accepted")
	}
}

// TestFASTReplyIdentityComesFromFinished checks validateKDCReplyIdentity is fed
// the identity in finished: a finished naming another client fails the reply
// even when the unprotected outer cname is the expected one.
func TestFASTReplyIdentityComesFromFinished(t *testing.T) {
	armorKey := bytes.Repeat([]byte{0x42}, 32)
	ticketRaw := []byte{0x61, 0x03, 0x02, 0x01, 0x05}
	c := NewClient("alice", "CORP.LOCAL", "127.0.0.1")
	ticket := messages.Ticket{Realm: "CORP.LOCAL", SName: messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}}}

	resp := fastFinishedFor(t, ticketRaw, armorKey, "CORP.LOCAL", "mallory")
	realm, name, err := verifyFASTFinished("FAST AS-REP", resp, ticketRaw, armorKey)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.validateKDCReplyIdentity("FAST AS-REP", realm, name, ticket, ticket.Realm, ticket.SName); err == nil {
		t.Error("a finished naming another client passed the identity check")
	}
}
