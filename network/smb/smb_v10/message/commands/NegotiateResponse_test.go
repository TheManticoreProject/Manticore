package commands_test

import (
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
	// SecurityMode (1 byte) left as zero.
	// Capabilities (4 bytes) at offset 19 selects the SMB_Data layout: when
	// CAP_EXTENDED_SECURITY is set the data block carries ServerGUID + SecurityBlob.
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
