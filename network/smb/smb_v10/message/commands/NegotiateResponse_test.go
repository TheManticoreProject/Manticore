package commands_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
)

// buildNegotiateResponseParameters constructs the marshalled parameters section
// for an extended-security NegotiateResponse, namely:
//
//	WordCount (1 byte) = 17
//	Words (34 bytes, as big-endian uint16s so AddWordsFromBytesStream reads them
//	       back as the little-endian SMB parameters it expects)
//
// The parameters fields together are 34 bytes (17 words):
//
//	DialectIndex(2) + SecurityMode(1) + MaxMpxCount(2) + MaxNumberVcs(2) +
//	MaxBufferSize(4) + MaxRawSize(4) + SessionKey(4) + Capabilities(4) +
//	SystemTime(8) + ServerTimeZone(2) + ChallengeLength(1) = 34 bytes
func buildNegotiateResponseParameters(caps capabilities.Capabilities) []byte {
	// Raw little-endian parameter bytes.
	params := make([]byte, 34)
	// DialectIndex = 0
	binary.LittleEndian.PutUint16(params[0:2], 0)
	// Capabilities (4 bytes) at offset 19: DialectIndex(2)+SecurityMode(1)+
	// MaxMpxCount(2)+MaxNumberVcs(2)+MaxBufferSize(4)+MaxRawSize(4)+SessionKey(4).
	// The data-block layout is selected by the CAP_EXTENDED_SECURITY bit.
	binary.LittleEndian.PutUint32(params[19:23], uint32(caps))
	// Remaining fields left as zero; ChallengeLength at offset 33 is 0.

	// Repack as 17 big-endian uint16 words so that Parameters.Unmarshal
	// (which reads big-endian and then AddWordsFromBytesStream is the
	// inverse of the byte-to-word transform) round-trips to these bytes.
	out := make([]byte, 1+34)
	out[0] = 17
	for i := 0; i < 17; i++ {
		// Words are big-endian in the wire format per parameters.Marshal.
		out[1+i*2] = params[i*2]
		out[1+i*2+1] = params[i*2+1]
	}
	return out
}

// Test_NegotiateResponse_Unmarshal_ShortExtendedSecurityDataDoesNotPanic
// verifies that a truncated extended-security NegotiateResponse (data section
// shorter than 16 bytes) returns an error instead of panicking on an
// out-of-range slice access.
func Test_NegotiateResponse_Unmarshal_ShortExtendedSecurityDataDoesNotPanic(t *testing.T) {
	paramsSection := buildNegotiateResponseParameters(capabilities.CAP_EXTENDED_SECURITY)

	// Data section: ByteCount = 4 (less than the 16 bytes required for ServerGUID)
	dataSection := []byte{0x04, 0x00, 0xAA, 0xBB, 0xCC, 0xDD}

	marshalled := append([]byte{}, paramsSection...)
	marshalled = append(marshalled, dataSection...)

	resp := commands.NewNegotiateResponse()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NegotiateResponse.Unmarshal panicked on short data: %v", r)
		}
	}()

	_, err := resp.Unmarshal(marshalled)
	if err == nil {
		t.Fatal("expected error from NegotiateResponse.Unmarshal on truncated data, got nil")
	}
}

// Test_NegotiateResponse_RoundTrip_ExtendedSecurity verifies that an
// extended-security response (CAP_EXTENDED_SECURITY set) marshals the
// ServerGUID + SecurityBlob data layout and unmarshals it back to the same
// values. This guards against the Marshal/Unmarshal data branches being
// inverted or keyed off the wrong discriminator.
func Test_NegotiateResponse_RoundTrip_ExtendedSecurity(t *testing.T) {
	resp := commands.NewNegotiateResponse()
	resp.Capabilities = capabilities.CAP_EXTENDED_SECURITY

	guidBytes := []byte{
		0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
		0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF,
	}
	resp.ServerGUID.FromRawBytes(guidBytes)
	resp.SecurityBlob = []byte{0xDE, 0xAD, 0xBE, 0xEF}

	marshalled, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed := commands.NewNegotiateResponse()
	if _, err := parsed.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !bytes.Equal(parsed.ServerGUID.ToBytes(), guidBytes) {
		t.Errorf("ServerGUID mismatch: got %x, want %x", parsed.ServerGUID.ToBytes(), guidBytes)
	}
	if !bytes.Equal(parsed.SecurityBlob, resp.SecurityBlob) {
		t.Errorf("SecurityBlob mismatch: got %x, want %x", parsed.SecurityBlob, resp.SecurityBlob)
	}
	if parsed.ChallengeLength != 0 {
		t.Errorf("ChallengeLength must be 0 for extended security, got %d", parsed.ChallengeLength)
	}
}

// Test_NegotiateResponse_RoundTrip_NonExtendedSecurity verifies that a
// non-extended-security response (CAP_EXTENDED_SECURITY clear) marshals the
// Challenge + DomainName data layout and unmarshals the Challenge back.
func Test_NegotiateResponse_RoundTrip_NonExtendedSecurity(t *testing.T) {
	resp := commands.NewNegotiateResponse()
	resp.Capabilities = 0
	resp.Challenge = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	// DomainName as a null-terminated UTF-16LE string "WG".
	resp.DomainName = []byte{0x57, 0x00, 0x47, 0x00, 0x00, 0x00}

	marshalled, err := resp.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	parsed := commands.NewNegotiateResponse()
	if _, err := parsed.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if parsed.ChallengeLength != 8 {
		t.Errorf("ChallengeLength: got %d, want 8", parsed.ChallengeLength)
	}
	if !bytes.Equal(parsed.Challenge, resp.Challenge) {
		t.Errorf("Challenge mismatch: got %x, want %x", parsed.Challenge, resp.Challenge)
	}
}
