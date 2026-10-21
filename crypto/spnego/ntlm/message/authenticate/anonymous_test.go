package authenticate

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
)

// TestCreateAuthenticateMessageAnonymous pins the [MS-NLMP] 3.3 special case for
// anonymous authentication: an empty username and password produce an empty
// NtChallengeResponse, an LmChallengeResponse of Z(1), and the
// NTLMSSP_NEGOTIATE_ANONYMOUS flag, with no MIC.
func TestCreateAuthenticateMessageAnonymous(t *testing.T) {
	for _, keyExch := range []bool{false, true} {
		negFlags := flags.NTLMSSP_NEGOTIATE_UNICODE |
			flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
			flags.NTLMSSP_NEGOTIATE_TARGET_INFO |
			flags.NTLMSSP_NEGOTIATE_SIGN
		if keyExch {
			negFlags |= flags.NTLMSSP_NEGOTIATE_KEY_EXCH
		}

		msg, err := CreateAuthenticateMessage(newTestChallenge(negFlags), "", "", "", "WORKSTATION")
		if err != nil {
			t.Fatalf("keyExch=%v: CreateAuthenticateMessage: %v", keyExch, err)
		}

		if !msg.IsAnonymous() || !msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_ANONYMOUS) {
			t.Errorf("keyExch=%v: NTLMSSP_NEGOTIATE_ANONYMOUS not set (flags %#08x)", keyExch, uint32(msg.NegotiateFlags))
		}
		if len(msg.NtChallengeResponse) != 0 {
			t.Errorf("keyExch=%v: NtChallengeResponse = % x, want empty", keyExch, msg.NtChallengeResponse)
		}
		if !bytes.Equal(msg.LmChallengeResponse, []byte{0x00}) {
			t.Errorf("keyExch=%v: LmChallengeResponse = % x, want Z(1)", keyExch, msg.LmChallengeResponse)
		}
		if msg.NeedsMIC {
			t.Errorf("keyExch=%v: NeedsMIC set on an anonymous AUTHENTICATE", keyExch)
		}
		if len(msg.UserName) != 0 {
			t.Errorf("keyExch=%v: UserName = % x, want empty", keyExch, msg.UserName)
		}
		if len(msg.SessionKey) != 16 {
			t.Errorf("keyExch=%v: SessionKey length = %d, want 16", keyExch, len(msg.SessionKey))
		}
		if keyExch != msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_KEY_EXCH) {
			t.Errorf("keyExch=%v: NTLMSSP_NEGOTIATE_KEY_EXCH asserted = %v", keyExch, !keyExch)
		}
		if keyExch {
			// The exported key travels RC4-wrapped under KeyExchangeKey = Z(16).
			unwrapped, err := rc4k(make([]byte, 16), msg.EncryptedRandomSessionKey)
			if err != nil {
				t.Fatalf("rc4k: %v", err)
			}
			if !bytes.Equal(unwrapped, msg.SessionKey) {
				t.Errorf("EncryptedRandomSessionKey does not unwrap under Z(16) to the exported session key")
			}
		} else if len(msg.EncryptedRandomSessionKey) != 0 || !bytes.Equal(msg.SessionKey, make([]byte, 16)) {
			t.Errorf("without key exchange: EncryptedRandomSessionKey = % x, SessionKey = % x; want empty, Z(16)",
				msg.EncryptedRandomSessionKey, msg.SessionKey)
		}

		// The response fields must survive the wire: a 0-byte NT response and a
		// 1-byte LM response.
		raw, err := msg.Marshal()
		if err != nil {
			t.Fatalf("keyExch=%v: Marshal: %v", keyExch, err)
		}
		var got AuthenticateMessage
		if _, err := got.Unmarshal(raw); err != nil {
			t.Fatalf("keyExch=%v: Unmarshal: %v", keyExch, err)
		}
		if len(got.NtChallengeResponse) != 0 || !bytes.Equal(got.LmChallengeResponse, []byte{0x00}) {
			t.Errorf("keyExch=%v: round trip NT=% x LM=% x, want empty and Z(1)", keyExch, got.NtChallengeResponse, got.LmChallengeResponse)
		}
		if !got.IsAnonymous() {
			t.Errorf("keyExch=%v: round trip lost NTLMSSP_NEGOTIATE_ANONYMOUS", keyExch)
		}
	}
}

// TestCreateAuthenticateMessageEmptyPasswordIsNotAnonymous confirms that a named
// user with an empty password still computes a real NTLMv2 response: only an empty
// username and an empty password together mean anonymous.
func TestCreateAuthenticateMessageEmptyPasswordIsNotAnonymous(t *testing.T) {
	negFlags := flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY
	msg, err := CreateAuthenticateMessage(newTestChallenge(negFlags), "User", "", "Domain", "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}
	if msg.IsAnonymous() {
		t.Error("NTLMSSP_NEGOTIATE_ANONYMOUS set for a named user")
	}
	if len(msg.NtChallengeResponse) < 16 {
		t.Errorf("NtChallengeResponse too short for NTLMv2: %d bytes", len(msg.NtChallengeResponse))
	}
}
