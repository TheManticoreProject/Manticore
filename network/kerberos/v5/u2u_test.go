package kerberos

import (
	"bytes"
	"encoding/asn1"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// seqChild parses a SEQUENCE TLV and returns the first element with the given
// class/tag.
func seqChild(t *testing.T, seqTLV []byte, class, tag int) asn1.RawValue {
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
			t.Fatalf("iterate: %v", err)
		}
		if e.Class == class && e.Tag == tag {
			return e
		}
		rest = r
	}
	t.Fatalf("no element class=%d tag=%d", class, tag)
	return asn1.RawValue{}
}

func TestU2URequestShape(t *testing.T) {
	c := fakeTGTClient(t) // from export_test.go: a client with a populated TGT

	// The target user's TGT (any APPLICATION[1] ticket for this structural test).
	targetTkt := messages.Ticket{
		TktVno:  messages.KerberosV5,
		Realm:   "CORP.LOCAL",
		SName:   messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", "CORP.LOCAL"}},
		EncPart: messages.EncryptedData{EType: messages.ETypeAES256CTSHMACSHA196, Cipher: bytes.Repeat([]byte{0x7a}, 16)},
	}
	targetTGTRaw, err := targetTkt.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	req, err := c.buildU2UTGSReq("victim", "", targetTGTRaw, 4242)
	if err != nil {
		t.Fatalf("buildU2UTGSReq: %v", err)
	}
	wire, err := req.Marshal()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// APPLICATION[12] -> inner SEQUENCE -> [4] body.
	var app asn1.RawValue
	if _, err := asn1.Unmarshal(wire, &app); err != nil {
		t.Fatal(err)
	}
	if app.Class != asn1.ClassApplication || app.Tag != messages.MsgTypeTGSReq {
		t.Fatalf("outer tag: class=%d tag=%d", app.Class, app.Tag)
	}
	bodyElem := seqChild(t, app.Bytes, asn1.ClassContextSpecific, 4)

	// KDCOptions [0]: BIT STRING; ENC-TKT-IN-SKEY is bit 28 -> byte 3, 0x08.
	optElem := seqChild(t, bodyElem.Bytes, asn1.ClassContextSpecific, 0)
	var flags asn1.BitString
	if _, err := asn1.Unmarshal(optElem.Bytes, &flags); err != nil {
		t.Fatalf("parse KDCOptions: %v", err)
	}
	if flags.At(kdcOptionEncTktInSKey) != 1 {
		t.Errorf("ENC-TKT-IN-SKEY (bit 28) not set; bytes=% X", flags.Bytes)
	}
	if flags.At(kdcOptionForwardable) != 1 {
		t.Errorf("forwardable (bit 1) not set")
	}

	// additional-tickets [11] -> SEQUENCE OF -> first is APPLICATION[1], and it
	// must be the target TGT verbatim.
	addl := seqChild(t, bodyElem.Bytes, asn1.ClassContextSpecific, 11)
	tkt := seqChild(t, addl.Bytes, asn1.ClassApplication, 1)
	if !bytes.Equal(tkt.FullBytes, targetTGTRaw) {
		t.Errorf("additional ticket is not the target TGT verbatim")
	}
}

func TestGetTGSU2UValidatesInputs(t *testing.T) {
	c := NewClient("svc", "corp.local", "10.0.0.1").WithPassword("x")
	if _, _, _, err := c.GetTGSU2U("victim", "", []byte{1}); err == nil {
		t.Error("expected error without a TGT")
	}
	c2 := fakeTGTClient(t)
	if _, _, _, err := c2.GetTGSU2U("", "", []byte{1}); err == nil {
		t.Error("expected error with empty target user")
	}
	if _, _, _, err := c2.GetTGSU2U("victim", "", nil); err == nil {
		t.Error("expected error with no target TGT")
	}
}
