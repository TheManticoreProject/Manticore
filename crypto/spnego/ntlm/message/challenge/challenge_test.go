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

// TestChallengeMessageRejectsAnotherMessageType asserts a NEGOTIATE is not parsed
// as a CHALLENGE. The three messages have different layouts and no discriminator
// beyond the header's MessageType, so parsing one as another fills the fields
// from the wrong offsets instead of failing.
func TestChallengeMessageRejectsAnotherMessageType(t *testing.T) {
	// A well-formed NEGOTIATE, long enough to clear the CHALLENGE length gate.
	message := make([]byte, 64)
	copy(message, header.NTLM_SIGNATURE[:])
	binary.LittleEndian.PutUint32(message[8:12], 1) // MESSAGE_TYPE_NEGOTIATE

	parsed := &challenge.ChallengeMessage{}
	if _, err := parsed.Unmarshal(message); err == nil {
		t.Error("Unmarshal() parsed a NEGOTIATE as a CHALLENGE")
	}
}
