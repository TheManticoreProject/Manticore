package gssapi

import (
	"bytes"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// buildForwardedCred builds a KRB-CRED carrying a single (unencrypted) forwarded
// TGT, as would be delegated over unconstrained delegation.
func buildForwardedCred(t *testing.T) ([]byte, *messages.KRBCred) {
	t.Helper()
	tkt := messages.Ticket{
		TktVno: messages.KerberosV5,
		Realm:  "CORP.LOCAL",
		SName:  messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}},
		EncPart: messages.EncryptedData{
			EType:  messages.ETypeAES256CTSHMACSHA196,
			Cipher: bytes.Repeat([]byte{0xCD}, 32),
		},
	}
	tktRaw, err := tkt.Marshal()
	if err != nil {
		t.Fatalf("Ticket.Marshal: %v", err)
	}
	enc := messages.EncKrbCredPart{
		TicketInfo: []messages.KrbCredInfo{{
			Key:       messages.EncryptionKey{KeyType: messages.ETypeAES256CTSHMACSHA196, KeyValue: bytes.Repeat([]byte{0x22}, 32)},
			PRealm:    "CORP.LOCAL",
			PName:     messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{"victim"}},
			SRealm:    "CORP.LOCAL",
			SName:     messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}},
			EndTime:   time.Date(2026, 7, 10, 20, 0, 0, 0, time.UTC),
			RenewTill: time.Date(2026, 7, 17, 20, 0, 0, 0, time.UTC),
		}},
	}
	encBytes, err := enc.Marshal()
	if err != nil {
		t.Fatalf("EncKrbCredPart.Marshal: %v", err)
	}
	cred := &messages.KRBCred{
		Tickets:    []messages.Ticket{tkt},
		TicketsRaw: [][]byte{tktRaw},
		EncPart:    messages.EncryptedData{EType: 0, Cipher: encBytes},
	}
	credBytes, err := cred.Marshal()
	if err != nil {
		t.Fatalf("KRBCred.Marshal: %v", err)
	}
	return credBytes, cred
}

// TestGSSChecksumDelegationRoundTrip builds a 0x8003 checksum with a delegated
// KRB-CRED and extracts it back, verifying the RFC 4121 Section 4.1.1 layout.
func TestGSSChecksumDelegationRoundTrip(t *testing.T) {
	credBytes, _ := buildForwardedCred(t)

	cksum := GSSChecksumValueWithDelegation(GSSMutualFlag, nil, credBytes)

	// The delegation flag must be set even though only GSSMutualFlag was passed.
	if cksum[20]&GSSDelegFlag == 0 {
		t.Errorf("delegation flag not set in checksum flags: % X", cksum[20:24])
	}

	got, err := ExtractDelegation(cksum)
	if err != nil {
		t.Fatalf("ExtractDelegation: %v", err)
	}
	if !bytes.Equal(got, credBytes) {
		t.Errorf("extracted KRB-CRED bytes differ from input")
	}
}

// TestExtractDelegationNoFlag verifies a checksum without the delegation flag
// yields no credential (and no error).
func TestExtractDelegationNoFlag(t *testing.T) {
	cksum := GSSChecksumValue(GSSMutualFlag, nil) // no deleg flag
	got, err := ExtractDelegation(cksum)
	if err != nil {
		t.Fatalf("ExtractDelegation: %v", err)
	}
	if got != nil {
		t.Errorf("expected no delegation, got %d bytes", len(got))
	}
}

// TestExtractDelegationTruncated verifies malformed checksums are rejected.
func TestExtractDelegationTruncated(t *testing.T) {
	if _, err := ExtractDelegation([]byte{1, 2, 3}); err == nil {
		t.Error("expected error for short checksum")
	}
	// Deleg flag set but no DlgOpt/Dlgth trailer.
	short := GSSChecksumValue(GSSDelegFlag, nil)
	if _, err := ExtractDelegation(short); err == nil {
		t.Error("expected error for deleg flag with no trailer")
	}
	// Dlgth claims more bytes than are present.
	bad := GSSChecksumValueWithDelegation(0, nil, []byte{0xAA, 0xBB, 0xCC, 0xDD})
	bad = bad[:len(bad)-2] // chop the KRB-CRED payload
	if _, err := ExtractDelegation(bad); err == nil {
		t.Error("expected error for Dlgth overrun")
	}
}

// TestExtractDelegatedCred exercises the full authenticator -> KRB-CRED ->
// EncKrbCredPart path against an unencrypted (etype 0) forwarded credential.
func TestExtractDelegatedCred(t *testing.T) {
	credBytes, want := buildForwardedCred(t)

	auth := &messages.Authenticator{
		AVno:   messages.KerberosV5,
		CRealm: "CORP.LOCAL",
		CName:  messages.PrincipalName{NameType: messages.NameTypePrincipal, NameString: []string{"victim"}},
		Cksum: &messages.Checksum{
			CKSumType: ChecksumTypeGSSAPI,
			Checksum:  GSSChecksumValueWithDelegation(GSSMutualFlag, nil, credBytes),
		},
		CTime: time.Now().UTC(),
	}

	cred, err := ExtractDelegatedCred(auth)
	if err != nil {
		t.Fatalf("ExtractDelegatedCred: %v", err)
	}
	if cred == nil {
		t.Fatal("expected a delegated credential")
	}
	if len(cred.Tickets) != 1 || cred.Tickets[0].Realm != want.Tickets[0].Realm {
		t.Errorf("delegated ticket mismatch: %+v", cred.Tickets)
	}

	// etype 0 enc-part: decrypt with an empty key (no crypto needed).
	part, err := DecryptDelegatedCredPart(cred, messages.EncryptionKey{})
	if err != nil {
		t.Fatalf("DecryptDelegatedCredPart: %v", err)
	}
	if len(part.TicketInfo) != 1 || part.TicketInfo[0].PName.NameString[0] != "victim" {
		t.Errorf("forwarded enc-part mismatch: %+v", part.TicketInfo)
	}
	wantEnd := time.Date(2026, 7, 10, 20, 0, 0, 0, time.UTC)
	if !part.TicketInfo[0].EndTime.Equal(wantEnd) {
		t.Errorf("forwarded ticket endtime: got %v want %v", part.TicketInfo[0].EndTime, wantEnd)
	}
}

// TestExtractDelegatedCredNonGSS verifies a non-GSS authenticator checksum is
// ignored (no delegation, no error).
func TestExtractDelegatedCredNonGSS(t *testing.T) {
	auth := &messages.Authenticator{
		Cksum: &messages.Checksum{CKSumType: 1, Checksum: []byte{1, 2, 3, 4}},
	}
	cred, err := ExtractDelegatedCred(auth)
	if err != nil || cred != nil {
		t.Errorf("expected no cred/no error for non-GSS checksum, got %v / %v", cred, err)
	}
}
