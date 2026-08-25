package client

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/signing"
)

// The signature algorithm itself is covered in the signing package; these tests
// cover the client's use of it: which sequence number is consumed, when signing
// applies at all, and what the connection state looks like afterwards.

// makeMessage returns a deterministic pseudo-message of the given length whose
// SecurityFeatures field starts zeroed, as a freshly marshalled SMB message would.
func makeMessage(n int) []byte {
	msg := make([]byte, n)
	for i := range msg {
		msg[i] = byte(i)
	}
	for i := signing.SecuritySignatureOffset; i < signing.SecuritySignatureOffset+signing.SecuritySignatureSize && i < n; i++ {
		msg[i] = 0
	}
	return msg
}

// TestSignOutgoingInactiveIsNoOp verifies that when signing is inactive the message
// is left untouched and no sequence number is consumed.
func TestSignOutgoingInactiveIsNoOp(t *testing.T) {
	c := &Client{Connection: &Connection{IsSigningActive: false}}
	msg := makeMessage(48)
	orig := append([]byte{}, msg...)

	respSeq, signed := c.signOutgoing(msg)
	if signed || respSeq != 0 {
		t.Errorf("expected (0, false) when inactive, got (%d, %v)", respSeq, signed)
	}
	if !bytes.Equal(msg, orig) {
		t.Error("message was modified while signing was inactive")
	}
	if err := c.verifyIncoming(msg, 0); err != nil {
		t.Errorf("verifyIncoming should be a no-op when inactive, got %v", err)
	}
}

// TestSignOutgoingActiveAdvancesSequence verifies that active signing signs the
// message, returns the response sequence number, and advances the send counter by
// two, and that the produced signature validates at the consumed sequence number.
func TestSignOutgoingActiveAdvancesSequence(t *testing.T) {
	key := bytes.Repeat([]byte{0xC3}, 16)
	c := &Client{Connection: &Connection{
		IsSigningActive:              true,
		SigningSessionKey:            key,
		ClientNextSendSequenceNumber: 2,
	}}

	msg := makeMessage(48)
	respSeq, signed := c.signOutgoing(msg)
	if !signed {
		t.Fatal("expected the message to be signed when active")
	}
	if respSeq != 3 {
		t.Errorf("expected response sequence 3, got %d", respSeq)
	}
	if c.Connection.ClientNextSendSequenceNumber != 4 {
		t.Errorf("expected next send sequence 4, got %d", c.Connection.ClientNextSendSequenceNumber)
	}
	// The request was signed with sequence 2.
	if !signing.Verify(key, msg, 2) {
		t.Error("signed message does not validate at the consumed sequence number")
	}

	// verifyIncoming should accept a response signed at the returned sequence number.
	resp := makeMessage(48)
	signing.Sign(key, resp, respSeq)
	if err := c.verifyIncoming(resp, respSeq); err != nil {
		t.Errorf("verifyIncoming rejected a valid response: %v", err)
	}
	// And reject a response with a bad signature.
	resp[30] ^= 0xFF
	if err := c.verifyIncoming(resp, respSeq); err == nil {
		t.Error("verifyIncoming accepted a tampered response")
	}
}

// TestSignOutgoingHandlesNilConnection asserts a client with no connection is a
// no-op rather than a nil dereference, since the guard is what every send path
// relies on before signing is ever armed.
func TestSignOutgoingHandlesNilConnection(t *testing.T) {
	c := &Client{}
	msg := makeMessage(48)

	if respSeq, signed := c.signOutgoing(msg); signed || respSeq != 0 {
		t.Errorf("expected (0, false) with no connection, got (%d, %v)", respSeq, signed)
	}
	if err := c.verifyIncoming(msg, 0); err != nil {
		t.Errorf("verifyIncoming with no connection should be a no-op, got %v", err)
	}
}

// TestSignOutgoingConsumesSuccessiveNumbers asserts consecutive requests on one
// connection take successive even numbers, and each expects the odd one above.
func TestSignOutgoingConsumesSuccessiveNumbers(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 16)
	c := &Client{Connection: &Connection{
		IsSigningActive:   true,
		SigningSessionKey: key,
		// Signing is armed by the AUTHENTICATE exchange, which consumes 0 and 1.
		ClientNextSendSequenceNumber: 2,
	}}

	for expected := uint32(2); expected <= 8; expected += 2 {
		msg := makeMessage(48)
		respSeq, signed := c.signOutgoing(msg)
		if !signed {
			t.Fatalf("request at %d was not signed", expected)
		}
		if !signing.Verify(key, msg, expected) {
			t.Fatalf("request was not signed at %d", expected)
		}
		if respSeq != expected+1 {
			t.Fatalf("expected response sequence %d, got %d", expected+1, respSeq)
		}
	}
	if c.Connection.ClientNextSendSequenceNumber != 10 {
		t.Fatalf("next send sequence = %d, want 10", c.Connection.ClientNextSendSequenceNumber)
	}
}
