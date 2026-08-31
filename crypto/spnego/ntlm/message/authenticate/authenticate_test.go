package authenticate_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// TestDomainNamePreservesCase guards against re-introducing the bug where the
// AUTHENTICATE DomainName field was upper-cased. NTLMv2 folds the domain into NTOWFv2
// with its original case (only the username is upper-cased, MS-NLMP 3.3.2), so the
// DomainName field MUST carry the same case; upper-casing it made the server compute a
// different NTProofStr and reject mixed/lower-case domains (e.g. an FQDN) with
// STATUS_LOGON_FAILURE.
func TestDomainNamePreservesCase(t *testing.T) {
	challengeMsg := &challenge.ChallengeMessage{
		NegotiateFlags: flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY,
	}
	copy(challengeMsg.ServerChallenge[:], []byte{1, 2, 3, 4, 5, 6, 7, 8})

	const domain = "TMP-W-2016.local" // mixed case, as an FQDN is typically supplied
	authMsg, err := authenticate.CreateAuthenticateMessage(challengeMsg, "Administrator", "pass", domain, "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}

	want := utf16.EncodeUTF16LE(domain)
	if !bytes.Equal(authMsg.DomainName, want) {
		t.Errorf("DomainName field was altered (case not preserved)\n got %q\nwant %q",
			utf16.DecodeUTF16LE(authMsg.DomainName), domain)
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	// Create a challenge message with some flags
	challengeMsg := &challenge.ChallengeMessage{
		NegotiateFlags: flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY,
	}
	copy(challengeMsg.ServerChallenge[:], []byte{1, 2, 3, 4, 5, 6, 7, 8})

	// Create an authenticate message
	authMsg, err := authenticate.CreateAuthenticateMessage(challengeMsg, "testuser", "testpass", "testdomain", "testworkstation")
	if err != nil {
		t.Fatalf("Failed to create authenticate message: %v", err)
	}

	authMsg.NegotiateFlags = challengeMsg.NegotiateFlags

	// Marshal the message
	data, err := authMsg.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Create a new message and unmarshal into it
	unmarshaledMsg := &authenticate.AuthenticateMessage{}
	_, err = unmarshaledMsg.Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Compare all fields
	if !bytes.Equal(authMsg.LmChallengeResponse, unmarshaledMsg.LmChallengeResponse) {
		t.Error("LmChallengeResponse mismatch")
		t.Errorf("Expected : %v", authMsg.LmChallengeResponse)
		t.Errorf("Actual   : %v", unmarshaledMsg.LmChallengeResponse)
	}
	if !bytes.Equal(authMsg.NtChallengeResponse, unmarshaledMsg.NtChallengeResponse) {
		t.Error("NtChallengeResponse mismatch")
		t.Errorf("Expected : %v", authMsg.NtChallengeResponse)
		t.Errorf("Actual   : %v", unmarshaledMsg.NtChallengeResponse)
	}
	if !bytes.Equal(authMsg.DomainName, unmarshaledMsg.DomainName) {
		t.Error("DomainName mismatch")
		t.Errorf("Expected : %v", authMsg.DomainName)
		t.Errorf("Actual   : %v", unmarshaledMsg.DomainName)
	}
	if !bytes.Equal(authMsg.UserName, unmarshaledMsg.UserName) {
		t.Error("UserName mismatch")
		t.Errorf("Expected : %v", authMsg.UserName)
		t.Errorf("Actual   : %v", unmarshaledMsg.UserName)
	}
	if !bytes.Equal(authMsg.Workstation, unmarshaledMsg.Workstation) {
		t.Error("Workstation mismatch")
		t.Errorf("Expected : %v", authMsg.Workstation)
		t.Errorf("Actual   : %v", unmarshaledMsg.Workstation)
	}
	if !bytes.Equal(authMsg.EncryptedRandomSessionKey, unmarshaledMsg.EncryptedRandomSessionKey) {
		t.Error("EncryptedRandomSessionKey mismatch")
		t.Errorf("Expected : %v", authMsg.EncryptedRandomSessionKey)
		t.Errorf("Actual   : %v", unmarshaledMsg.EncryptedRandomSessionKey)
	}
	if authMsg.NegotiateFlags != unmarshaledMsg.NegotiateFlags {
		t.Error("NegotiateFlags mismatch")
		t.Errorf("Expected : %v", authMsg.NegotiateFlags)
		t.Errorf("Actual   : %v", unmarshaledMsg.NegotiateFlags)
	}

	// Compare DataFields structures
	if authMsg.LmChallengeResponseFields != unmarshaledMsg.LmChallengeResponseFields {
		t.Error("LmChallengeResponseFields mismatch")
		t.Errorf("Expected : %v", authMsg.LmChallengeResponseFields)
		t.Errorf("Actual   : %v", unmarshaledMsg.LmChallengeResponseFields)
	}
	if authMsg.NtChallengeResponseFields != unmarshaledMsg.NtChallengeResponseFields {
		t.Error("NtChallengeResponseFields mismatch")
		t.Errorf("Expected : %v", authMsg.NtChallengeResponseFields)
		t.Errorf("Actual   : %v", unmarshaledMsg.NtChallengeResponseFields)
	}
	if authMsg.DomainNameFields != unmarshaledMsg.DomainNameFields {
		t.Error("DomainNameFields mismatch")
		t.Errorf("Expected : %v", authMsg.DomainNameFields)
		t.Errorf("Actual   : %v", unmarshaledMsg.DomainNameFields)
	}
	if authMsg.UserNameFields != unmarshaledMsg.UserNameFields {
		t.Error("UserNameFields mismatch")
		t.Errorf("Expected : %v", authMsg.UserNameFields)
		t.Errorf("Actual   : %v", unmarshaledMsg.UserNameFields)
	}
	if authMsg.WorkstationFields != unmarshaledMsg.WorkstationFields {
		t.Error("WorkstationFields mismatch")
		t.Errorf("Expected : %v", authMsg.WorkstationFields)
		t.Errorf("Actual   : %v", unmarshaledMsg.WorkstationFields)
	}
	if authMsg.EncryptedRandomSessionKeyFields != unmarshaledMsg.EncryptedRandomSessionKeyFields {
		t.Error("EncryptedRandomSessionKeyFields mismatch")
		t.Errorf("Expected : %v", authMsg.EncryptedRandomSessionKeyFields)
		t.Errorf("Actual   : %v", unmarshaledMsg.EncryptedRandomSessionKeyFields)
	}
}

// TestAuthenticateMessagePayloadOffsetOverflow asserts that a payload
// BufferOffset near the top of the 32-bit range is refused rather than becoming a
// negative index that passes the bounds check and panics in the slice expression.
//
// The check was written as int(BufferOffset)+int(Len) > len(data). int is 64 bits
// on the hosts this is usually run on, where that is correct; it is 32 bits on the
// 386 targets this project also builds and tests, where int(0xFFFFFFFF) is -1 and
// the comparison passes for an offset nowhere near the buffer.
func TestAuthenticateMessagePayloadOffsetOverflow(t *testing.T) {
	// The fixed part of an AUTHENTICATE is 88 bytes at minimum: signature and
	// type (12), six field descriptors (48), NegotiateFlags (4), Version (8).
	// Each subtest sets one descriptor's Len and drives its offset to the top of
	// the range, leaving the others empty so only the field under test is read.
	descriptors := map[string]int{
		"LmChallengeResponse":       12,
		"NtChallengeResponse":       20,
		"DomainName":                28,
		"UserName":                  36,
		"Workstation":               44,
		"EncryptedRandomSessionKey": 52,
	}

	for name, at := range descriptors {
		t.Run(name, func(t *testing.T) {
			message := make([]byte, 96)
			copy(message, []byte{'N', 'T', 'L', 'M', 'S', 'S', 'P', 0})
			binary.LittleEndian.PutUint32(message[8:12], 3)
			binary.LittleEndian.PutUint16(message[at:at+2], 8)   // Len
			binary.LittleEndian.PutUint16(message[at+2:at+4], 8) // MaxLen
			binary.LittleEndian.PutUint32(message[at+4:at+8], 0xFFFFFFFF)

			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("Unmarshal() panicked instead of returning an error: %v", recovered)
				}
			}()

			parsed := &authenticate.AuthenticateMessage{}
			if _, err := parsed.Unmarshal(message); err == nil {
				t.Error("Unmarshal() accepted a payload offset outside the message")
			}
		})
	}
}
