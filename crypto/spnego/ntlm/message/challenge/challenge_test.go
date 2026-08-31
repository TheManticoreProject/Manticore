package challenge_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/header"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
)

func TestChallengeMessageMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		message *challenge.ChallengeMessage
	}{
		{
			name: "Basic challenge message",
			message: &challenge.ChallengeMessage{
				Header: header.Header{
					Signature:   [8]byte{'N', 'T', 'L', 'M', 'S', 'S', 'P', 0},
					MessageType: 2,
				},
				NegotiateFlags:  flags.NTLMSSP_NEGOTIATE_UNICODE,
				ServerChallenge: [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
				Reserved:        [8]byte{},
				TargetName:      []byte("DOMAIN"),
				TargetInfo:      []byte{0x01, 0x02, 0x03, 0x04},
			},
		},
		{
			name: "Challenge message with version",
			message: &challenge.ChallengeMessage{
				Header: header.Header{
					Signature:   [8]byte{'N', 'T', 'L', 'M', 'S', 'S', 'P', 0},
					MessageType: 2,
				},
				NegotiateFlags:  flags.NTLMSSP_NEGOTIATE_VERSION,
				ServerChallenge: [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
				Reserved:        [8]byte{},
				TargetName:      []byte("DOMAIN"),
				TargetInfo:      []byte{0x01, 0x02, 0x03, 0x04},
				Version: &version.Version{
					ProductMajorVersion: 6,
					ProductMinorVersion: 1,
					ProductBuild:        7601,
					NTLMRevision:        15,
				},
			},
		},
		{
			name: "Challenge message with multiple flags",
			message: &challenge.ChallengeMessage{
				Header: header.Header{
					Signature:   [8]byte{'N', 'T', 'L', 'M', 'S', 'S', 'P', 0},
					MessageType: 2,
				},
				NegotiateFlags:  flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_SIGN | flags.NTLMSSP_NEGOTIATE_SEAL,
				ServerChallenge: [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
				Reserved:        [8]byte{},
				TargetName:      []byte("DOMAIN"),
				TargetInfo:      []byte{0x01, 0x02, 0x03, 0x04},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal the message
			data, err := tt.message.Marshal()
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			// Create a new message and unmarshal into it
			unmarshaledMsg := &challenge.ChallengeMessage{}
			_, err = unmarshaledMsg.Unmarshal(data)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			// Compare fields
			if !bytes.Equal(tt.message.TargetName, unmarshaledMsg.TargetName) {
				t.Error("TargetName mismatch")
				t.Errorf("Expected : %s", hex.EncodeToString(tt.message.TargetName))
				t.Errorf("Actual   : %s", hex.EncodeToString(unmarshaledMsg.TargetName))
			}
			if !bytes.Equal(tt.message.TargetInfo, unmarshaledMsg.TargetInfo) {
				t.Error("TargetInfo mismatch")
				t.Errorf("Expected : %s", hex.EncodeToString(tt.message.TargetInfo))
				t.Errorf("Actual   : %s", hex.EncodeToString(unmarshaledMsg.TargetInfo))
			}
			if tt.message.NegotiateFlags != unmarshaledMsg.NegotiateFlags {
				t.Error("NegotiateFlags mismatch")
				t.Errorf("Expected : %v", tt.message.NegotiateFlags)
				t.Errorf("Actual   : %v", unmarshaledMsg.NegotiateFlags)
			}
			if tt.message.ServerChallenge != unmarshaledMsg.ServerChallenge {
				t.Error("ServerChallenge mismatch")
				t.Errorf("Expected : %v", tt.message.ServerChallenge)
				t.Errorf("Actual   : %v", unmarshaledMsg.ServerChallenge)
			}
			if tt.message.Reserved != unmarshaledMsg.Reserved {
				t.Error("Reserved mismatch")
				t.Errorf("Expected : %v", tt.message.Reserved)
				t.Errorf("Actual   : %v", unmarshaledMsg.Reserved)
			}
			if tt.message.TargetNameFields != unmarshaledMsg.TargetNameFields {
				t.Error("TargetNameFields mismatch")
				t.Errorf("Expected : %v", tt.message.TargetNameFields)
				t.Errorf("Actual   : %v", unmarshaledMsg.TargetNameFields)
			}

			// Compare Version if present
			if tt.message.Version != nil {
				if unmarshaledMsg.Version == nil {
					t.Error("Version missing in unmarshaled message")
				} else {
					if *tt.message.Version != *unmarshaledMsg.Version {
						t.Error("Version mismatch")
					}
				}
			}
		})
	}
}

// TestChallengeMessagePayloadOffsetOverflow asserts that a TargetName or
// TargetInfo BufferOffset near the top of the 32-bit range is refused rather than
// wrapping past the bounds check and panicking in the slice expression.
//
// BufferOffset is a 32-bit wire field and Len a 16-bit one, so computing the end
// of the payload as BufferOffset+Len in uint32 overflows: 0xFFFFFFFF plus a small
// length compares as inside the buffer.
func TestChallengeMessagePayloadOffsetOverflow(t *testing.T) {
	// The fixed part of a CHALLENGE: signature and type (12), TargetNameFields
	// (8), NegotiateFlags (4), ServerChallenge (8), Reserved (8),
	// TargetInfoFields (8), Version (8).
	build := func(targetNameOffset, targetInfoOffset uint32) []byte {
		message := make([]byte, 60)
		copy(message, []byte{'N', 'T', 'L', 'M', 'S', 'S', 'P', 0})
		binary.LittleEndian.PutUint32(message[8:12], 2)

		binary.LittleEndian.PutUint16(message[12:14], 8) // TargetName Len
		binary.LittleEndian.PutUint16(message[14:16], 8) // TargetName MaxLen
		binary.LittleEndian.PutUint32(message[16:20], targetNameOffset)

		binary.LittleEndian.PutUint16(message[40:42], 8) // TargetInfo Len
		binary.LittleEndian.PutUint16(message[42:44], 8) // TargetInfo MaxLen
		binary.LittleEndian.PutUint32(message[44:48], targetInfoOffset)

		return message
	}

	tests := []struct {
		name    string
		message []byte
	}{
		{name: "TargetName offset overflows", message: build(0xFFFFFFFF, 52)},
		{name: "TargetInfo offset overflows", message: build(52, 0xFFFFFFFF)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("Unmarshal() panicked instead of returning an error: %v", recovered)
				}
			}()

			parsed := &challenge.ChallengeMessage{}
			if _, err := parsed.Unmarshal(test.message); err == nil {
				t.Error("Unmarshal() accepted a payload offset outside the message")
			}
		})
	}
}
