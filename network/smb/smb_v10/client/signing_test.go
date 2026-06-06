package client

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/securityfeatures"
)

// TestSecuritySignatureOffsetMatchesHeader guards the securitySignatureOffset
// constant against drift in the SMB header layout: a marshalled header MUST carry
// its 8-byte SecurityFeatures field exactly at securitySignatureOffset.
func TestSecuritySignatureOffsetMatchesHeader(t *testing.T) {
	h := header.NewHeader()
	sig := securityfeatures.NewSecurityFeaturesSecuritySignature()
	sig.SetSecuritySignature([8]byte{0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA})
	h.SecurityFeatures = sig

	raw, err := h.Marshal()
	if err != nil {
		t.Fatalf("header Marshal failed: %v", err)
	}
	if len(raw) != header.SMB_HEADER_SIZE {
		t.Fatalf("header is %d bytes, want %d", len(raw), header.SMB_HEADER_SIZE)
	}
	want := bytes.Repeat([]byte{0xAA}, securitySignatureSize)
	if !bytes.Equal(raw[securitySignatureOffset:securitySignatureOffset+securitySignatureSize], want) {
		t.Errorf("SecurityFeatures not at offset %d: bytes there = % x", securitySignatureOffset, raw[securitySignatureOffset:securitySignatureOffset+securitySignatureSize])
	}
}

// makeMessage returns a deterministic pseudo-message of the given length whose
// SecurityFeatures field (bytes 14..22) starts zeroed, as a freshly marshalled SMB
// message would.
func makeMessage(n int) []byte {
	msg := make([]byte, n)
	for i := range msg {
		msg[i] = byte(i)
	}
	for i := securitySignatureOffset; i < securitySignatureOffset+securitySignatureSize && i < n; i++ {
		msg[i] = 0
	}
	return msg
}

// TestComputeSMBSignatureMatchesFormula pins the signature algorithm: the 8-byte
// signature MUST equal the first 8 bytes of MD5(MACKey || message), with the
// message's signature field set to the little-endian sequence number followed by
// four zero bytes. The independent computation here guards the field offset, the
// sequence-number placement, the endianness, and the digest input order.
func TestComputeSMBSignatureMatchesFormula(t *testing.T) {
	key := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	msg := makeMessage(64)
	const seq = uint32(0x04030201)

	// Independent reference computation.
	ref := make([]byte, len(msg))
	copy(ref, msg)
	binary.LittleEndian.PutUint32(ref[securitySignatureOffset:securitySignatureOffset+4], seq)
	for i := securitySignatureOffset + 4; i < securitySignatureOffset+8; i++ {
		ref[i] = 0
	}
	sum := md5.Sum(append(append([]byte{}, key...), ref...))
	var want [8]byte
	copy(want[:], sum[:8])

	got := computeSMBSignature(key, msg, seq)
	if got != want {
		t.Errorf("signature = % x, want % x", got, want)
	}

	// The input message must not have been modified.
	if msg[securitySignatureOffset] != 0 {
		t.Error("computeSMBSignature modified the caller's buffer")
	}
}

// TestSignAndVerifyRoundTrip verifies a signed message validates with the same key
// and sequence number, and that the signature is written at the right offset.
func TestSignAndVerifyRoundTrip(t *testing.T) {
	key := bytes.Repeat([]byte{0xAB}, 16)
	msg := makeMessage(48)
	const seq = uint32(7)

	signSMBMessage(key, msg, seq)

	if !verifySMBSignature(key, msg, seq) {
		t.Fatal("verifySMBSignature failed on a freshly signed message")
	}
	// The signature must have been written into the SecurityFeatures field.
	if bytes.Equal(msg[securitySignatureOffset:securitySignatureOffset+8], make([]byte, 8)) {
		t.Error("signature field is still zero after signing")
	}
}

// TestVerifyRejectsTamperingAndWrongSequence verifies that signature validation
// fails on a modified body or a mismatched sequence number.
func TestVerifyRejectsTamperingAndWrongSequence(t *testing.T) {
	key := bytes.Repeat([]byte{0x5A}, 16)
	const seq = uint32(2)

	msg := makeMessage(48)
	signSMBMessage(key, msg, seq)

	// Wrong sequence number.
	if verifySMBSignature(key, msg, seq+1) {
		t.Error("verification passed with the wrong sequence number")
	}

	// Tampered payload (a byte outside the signature field).
	tampered := append([]byte{}, msg...)
	tampered[40] ^= 0xFF
	if verifySMBSignature(key, tampered, seq) {
		t.Error("verification passed on a tampered message")
	}

	// Wrong key.
	if verifySMBSignature(bytes.Repeat([]byte{0x00}, 16), msg, seq) {
		t.Error("verification passed with the wrong key")
	}
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
	if !verifySMBSignature(key, msg, 2) {
		t.Error("signed message does not validate at the consumed sequence number")
	}

	// verifyIncoming should accept a response signed at the returned sequence number.
	resp := makeMessage(48)
	signSMBMessage(key, resp, respSeq)
	if err := c.verifyIncoming(resp, respSeq); err != nil {
		t.Errorf("verifyIncoming rejected a valid response: %v", err)
	}
	// And reject a response with a bad signature.
	resp[30] ^= 0xFF
	if err := c.verifyIncoming(resp, respSeq); err == nil {
		t.Error("verifyIncoming accepted a tampered response")
	}
}
