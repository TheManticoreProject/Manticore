package messages

import (
	"bytes"
	"encoding/asn1"
	"testing"
	"time"
)

// childByTag parses a SEQUENCE TLV and returns the first element whose
// class/tag match.
func childByTag(t *testing.T, seqTLV []byte, class, tag int) asn1.RawValue {
	t.Helper()
	var seq asn1.RawValue
	if _, err := asn1.Unmarshal(seqTLV, &seq); err != nil {
		t.Fatalf("parse SEQUENCE: %v", err)
	}
	rest := seq.Bytes
	for len(rest) > 0 {
		var e asn1.RawValue
		r, err := asn1.Unmarshal(rest, &e)
		if err != nil {
			t.Fatalf("iterate SEQUENCE: %v", err)
		}
		if e.Class == class && e.Tag == tag {
			return e
		}
		rest = r
	}
	t.Fatalf("no element with class=%d tag=%d", class, tag)
	return asn1.RawValue{}
}

func sampleTicket(t *testing.T) (Ticket, []byte) {
	t.Helper()
	tkt := Ticket{
		TktVno:  KerberosV5,
		Realm:   "CORP.LOCAL",
		SName:   PrincipalName{NameType: NameTypeSRVInst, NameString: []string{"cifs", "dc01.corp.local"}},
		EncPart: EncryptedData{EType: ETypeAES256CTSHMACSHA196, Cipher: bytes.Repeat([]byte{0x5a}, 16)},
	}
	raw, err := tkt.Marshal()
	if err != nil {
		t.Fatalf("Ticket.Marshal: %v", err)
	}
	return tkt, raw
}

// TestKDCReqBodyAdditionalTicketsEncoding verifies that additional-tickets are
// spliced in as [11] EXPLICIT SEQUENCE OF Ticket with each ticket kept in its
// APPLICATION[1] form and fully recoverable.
func TestKDCReqBodyAdditionalTicketsEncoding(t *testing.T) {
	_, tktRaw := sampleTicket(t)

	body := KDCReqBody{
		KDCOptions:      NewKerberosFlags(1),
		Realm:           "CORP.LOCAL",
		SName:           PrincipalName{NameType: NameTypeSRVInst, NameString: []string{"host", "target"}},
		Till:            time.Date(2026, 7, 9, 20, 0, 0, 0, time.UTC),
		Nonce:           12345,
		EType:           []int{ETypeAES256CTSHMACSHA196},
		AdditTicketsRaw: [][]byte{tktRaw},
	}
	bodyTLV, err := encodeKDCReqBodyForTGS(body)
	if err != nil {
		t.Fatalf("encodeKDCReqBodyForTGS: %v", err)
	}

	// body -> [11] -> SEQUENCE OF -> first element must be APPLICATION[1].
	// elem11.Bytes is the SEQUENCE-OF Ticket TLV (the [11] EXPLICIT content).
	elem11 := childByTag(t, bodyTLV, asn1.ClassContextSpecific, 11)
	first := childByTag(t, elem11.Bytes, asn1.ClassApplication, 1)
	if !bytes.Equal(first.FullBytes, tktRaw) {
		t.Errorf("additional ticket bytes not preserved verbatim")
	}
	// And it must reparse as a Ticket.
	var got Ticket
	if _, err := got.Unmarshal(first.FullBytes); err != nil {
		t.Fatalf("reparse additional ticket: %v", err)
	}
	if got.Realm != "CORP.LOCAL" || got.SName.NameString[0] != "cifs" {
		t.Errorf("additional ticket fields wrong: %+v", got)
	}
}

// TestKDCReqBodyNoAdditionalTickets confirms the body is unchanged (no [11])
// when there are no additional tickets.
func TestKDCReqBodyNoAdditionalTickets(t *testing.T) {
	body := KDCReqBody{
		KDCOptions: NewKerberosFlags(1),
		Realm:      "CORP.LOCAL",
		SName:      PrincipalName{NameType: NameTypeSRVInst, NameString: []string{"host", "target"}},
		Till:       time.Date(2026, 7, 9, 20, 0, 0, 0, time.UTC),
		Nonce:      1,
		EType:      []int{ETypeAES256CTSHMACSHA196},
	}
	withHelper, err := encodeKDCReqBodyForTGS(body)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := asn1.Marshal(marshalKDCReqBody(body))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(withHelper, plain) {
		t.Error("body with no additional tickets should equal the plain struct marshal")
	}
}
