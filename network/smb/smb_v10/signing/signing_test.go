package signing

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/securityfeatures"
)

// TestSecuritySignatureOffsetMatchesHeader guards SecuritySignatureOffset against
// drift in the SMB header layout: a marshalled header MUST carry its 8-byte
// SecurityFeatures field exactly there.
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
	want := bytes.Repeat([]byte{0xAA}, SecuritySignatureSize)
	if !bytes.Equal(raw[SecuritySignatureOffset:SecuritySignatureOffset+SecuritySignatureSize], want) {
		t.Errorf("SecurityFeatures not at offset %d: bytes there = % x",
			SecuritySignatureOffset, raw[SecuritySignatureOffset:SecuritySignatureOffset+SecuritySignatureSize])
	}
}

// makeMessage returns a deterministic pseudo-message of the given length whose
// SecurityFeatures field starts zeroed, as a freshly marshalled SMB message would.
func makeMessage(n int) []byte {
	msg := make([]byte, n)
	for i := range msg {
		msg[i] = byte(i)
	}
	for i := SecuritySignatureOffset; i < SecuritySignatureOffset+SecuritySignatureSize && i < n; i++ {
		msg[i] = 0
	}
	return msg
}

// TestComputeMatchesFormula pins the algorithm: the signature MUST equal the first
// 8 bytes of MD5(MACKey || message), with the message's signature field set to the
// little-endian sequence number followed by four zero bytes. The independent
// computation here guards the field offset, the sequence-number placement, the
// endianness and the digest input order.
func TestComputeMatchesFormula(t *testing.T) {
	key := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	msg := makeMessage(64)
	const seq = uint32(0x04030201)

	// Independent reference computation.
	ref := make([]byte, len(msg))
	copy(ref, msg)
	binary.LittleEndian.PutUint32(ref[SecuritySignatureOffset:SecuritySignatureOffset+4], seq)
	for i := SecuritySignatureOffset + 4; i < SecuritySignatureOffset+8; i++ {
		ref[i] = 0
	}
	sum := md5.Sum(append(append([]byte{}, key...), ref...))
	var want Signature
	copy(want[:], sum[:8])

	got := Compute(key, msg, seq)
	if got != want {
		t.Errorf("signature = % x, want % x", got, want)
	}

	// The input message must not have been modified.
	if msg[SecuritySignatureOffset] != 0 {
		t.Error("Compute modified the caller's buffer")
	}
}

// TestSignAndVerifyRoundTrip verifies a signed message validates with the same key
// and sequence number, and that the signature is written at the right offset.
func TestSignAndVerifyRoundTrip(t *testing.T) {
	key := bytes.Repeat([]byte{0xAB}, 16)
	msg := makeMessage(48)
	const seq = uint32(7)

	Sign(key, msg, seq)

	if !Verify(key, msg, seq) {
		t.Fatal("Verify failed on a freshly signed message")
	}
	// The signature must have been written into the SecurityFeatures field.
	if bytes.Equal(msg[SecuritySignatureOffset:SecuritySignatureOffset+8], make([]byte, 8)) {
		t.Error("signature field is still zero after signing")
	}
}

// TestVerifyRejectsTamperingAndWrongSequence verifies that validation fails on a
// modified body, a mismatched sequence number, or the wrong key.
func TestVerifyRejectsTamperingAndWrongSequence(t *testing.T) {
	key := bytes.Repeat([]byte{0x5A}, 16)
	const seq = uint32(2)

	msg := makeMessage(48)
	Sign(key, msg, seq)

	if Verify(key, msg, seq+1) {
		t.Error("verification passed with the wrong sequence number")
	}

	tampered := append([]byte{}, msg...)
	tampered[40] ^= 0xFF
	if Verify(key, tampered, seq) {
		t.Error("verification passed on a tampered message")
	}

	if Verify(bytes.Repeat([]byte{0x00}, 16), msg, seq) {
		t.Error("verification passed with the wrong key")
	}
}

// TestShortMessages asserts a message too short to carry a signature is handled
// rather than indexed into. A signing implementation is reachable from the network
// on the server side, so a short frame must not panic.
func TestShortMessages(t *testing.T) {
	key := bytes.Repeat([]byte{0x11}, 16)

	for length := 0; length < SecuritySignatureOffset+SecuritySignatureSize; length++ {
		short := make([]byte, length)

		// Sign leaves it alone: there is no field to write into.
		before := append([]byte{}, short...)
		Sign(key, short, 1)
		if !bytes.Equal(short, before) {
			t.Fatalf("Sign modified a %d-byte message that cannot carry a signature", length)
		}

		// Verify reports failure rather than reading past the end.
		if Verify(key, short, 1) {
			t.Fatalf("Verify accepted a %d-byte message", length)
		}

		// Compute must not panic, and must not read past the end.
		Compute(key, short, 1)
	}
}

// TestSequenceNumberRule pins the pairing rule both sides depend on: a request
// takes a number, its response the one above, and the next request skips both.
// Getting this wrong in one direction produces signatures the other side rejects.
func TestSequenceNumberRule(t *testing.T) {
	// The AUTHENTICATE request is signed at 0, so its response is at 1 and the
	// first request after it is at 2.
	if got := ResponseSequenceNumber(0); got != 1 {
		t.Errorf("ResponseSequenceNumber(0) = %d, want 1", got)
	}
	if got := NextRequestSequenceNumber(0); got != 2 {
		t.Errorf("NextRequestSequenceNumber(0) = %d, want 2", got)
	}

	// Requests stay even and responses odd as the exchange advances.
	sequence := uint32(0)
	for exchange := 0; exchange < 8; exchange++ {
		if sequence%2 != 0 {
			t.Fatalf("request %d is signed at %d, which is not even", exchange, sequence)
		}
		if response := ResponseSequenceNumber(sequence); response%2 == 0 {
			t.Fatalf("response %d is signed at %d, which is not odd", exchange, response)
		}
		sequence = NextRequestSequenceNumber(sequence)
	}
	if sequence != 16 {
		t.Fatalf("after 8 exchanges the sequence is %d, want 16", sequence)
	}
}

// TestSignVerifyAcrossSides asserts the primitives are usable in both directions:
// what one party signs at a sequence number, the other verifies at the same one.
// That symmetry is why the package is direction-agnostic.
func TestSignVerifyAcrossSides(t *testing.T) {
	key := bytes.Repeat([]byte{0x7E}, 16)

	// A request signed by the sending side.
	request := makeMessage(64)
	requestSequence := uint32(4)
	Sign(key, request, requestSequence)

	// The receiving side verifies at the number it expects.
	if !Verify(key, request, requestSequence) {
		t.Fatal("the receiving side rejected a validly signed request")
	}

	// It then signs its reply at the number above, which the original sender
	// verifies at the value it was told to expect.
	response := makeMessage(48)
	responseSequence := ResponseSequenceNumber(requestSequence)
	Sign(key, response, responseSequence)
	if !Verify(key, response, responseSequence) {
		t.Fatal("the requesting side rejected a validly signed response")
	}
	// And the response must not verify at the request's number.
	if Verify(key, response, requestSequence) {
		t.Fatal("the response verified at the request's sequence number")
	}
}
