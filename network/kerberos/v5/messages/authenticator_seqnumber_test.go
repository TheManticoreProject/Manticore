package messages

import (
	"bytes"
	"testing"
)

// seqNumberElement is the [7] EXPLICIT INTEGER 0 encoding an authenticator must
// carry when a deliberate seq-number of 0 is requested: a7 03 02 01 00.
var seqNumberZeroElement = []byte{0xa7, 0x03, 0x02, 0x01, 0x00}

func baseAuthenticator() *Authenticator {
	return &Authenticator{
		AVno:   KerberosV5,
		CRealm: "CORP.LOCAL",
		CName:  cliName(),
		CUSec:  123,
		CTime:  tstTime,
	}
}

// TestAuthenticatorSeqNumberZeroEmittedWhenPresent guards the DCE/RPC per-message
// path: a requested seq-number of 0 must appear on the wire, not be dropped as an
// optional zero value by encoding/asn1.
func TestAuthenticatorSeqNumberZeroEmittedWhenPresent(t *testing.T) {
	a := baseAuthenticator()
	a.SeqNumber = 0
	a.SeqNumberPresent = true

	wire, err := a.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !bytes.Contains(wire, seqNumberZeroElement) {
		t.Fatalf("seq-number 0 not emitted; authenticator = %x", wire)
	}

	var got Authenticator
	if _, err := got.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.SeqNumber != 0 {
		t.Fatalf("decoded SeqNumber = %d, want 0", got.SeqNumber)
	}
}

// TestAuthenticatorSeqNumberOmittedWhenZeroAndNotPresent confirms the default
// behavior is unchanged: a zero seq-number with no presence flag is omitted.
func TestAuthenticatorSeqNumberOmittedWhenZeroAndNotPresent(t *testing.T) {
	a := baseAuthenticator()
	a.SeqNumber = 0
	a.SeqNumberPresent = false

	wire, err := a.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if bytes.Contains(wire, seqNumberZeroElement) {
		t.Fatalf("seq-number 0 wrongly emitted; authenticator = %x", wire)
	}
	// No [7] element at all when the value is zero and not forced.
	if idx := bytes.IndexByte(wire, 0xa7); idx != -1 {
		t.Fatalf("unexpected [7] element at offset %d; authenticator = %x", idx, wire)
	}
}

// TestAuthenticatorSeqNumberNonZeroRoundTrip confirms a non-zero seq-number is
// still emitted and preserved regardless of the presence flag.
func TestAuthenticatorSeqNumberNonZeroRoundTrip(t *testing.T) {
	a := baseAuthenticator()
	a.SeqNumber = 0x01020304

	wire, err := a.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got Authenticator
	if _, err := got.Unmarshal(wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.SeqNumber != 0x01020304 {
		t.Fatalf("decoded SeqNumber = %#x, want 0x01020304", got.SeqNumber)
	}
}
