package authenticate_test

import (
	"bytes"
	"crypto/rc4"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
)

// noESSChallenge builds a CHALLENGE that does not offer
// NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY, with the fixed server challenge a
// rogue server uses to make NTLMv1 responses crackable with rainbow tables.
func noESSChallenge(negFlags flags.NegotiateFlags) *challenge.ChallengeMessage {
	msg := &challenge.ChallengeMessage{NegotiateFlags: negFlags}
	copy(msg.ServerChallenge[:], []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88})
	return msg
}

// TestCreateAuthenticateMessageWithoutESSUsesNTLMv2 guards against the NTLMv1
// downgrade: a server that clears the extended session security bit must not be
// able to obtain an NTLMv1 (and LM) response. NTLMv2 does not depend on the server
// negotiating ESS ([MS-NLMP] 3.3.2), so the client always sends an NTLMv2
// response, together with a session key consistent with the flags it asserts.
func TestCreateAuthenticateMessageWithoutESSUsesNTLMv2(t *testing.T) {
	cases := []struct {
		name     string
		negFlags flags.NegotiateFlags
	}{
		{"key exchange", flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_NTLM | flags.NTLMSSP_NEGOTIATE_KEY_EXCH},
		{"no key exchange", flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_NTLM},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := authenticate.CreateAuthenticateMessage(noESSChallenge(tc.negFlags), kxUser, kxPassword, kxDomain, "WORKSTATION")
			if err != nil {
				t.Fatalf("CreateAuthenticateMessage: %v", err)
			}

			if len(msg.NtChallengeResponse) <= 24 {
				t.Fatalf("NtChallengeResponse is %d bytes, want an NTLMv2 response (> 24 bytes)", len(msg.NtChallengeResponse))
			}
			if len(msg.SessionKey) != 16 {
				t.Fatalf("SessionKey is %d bytes, want 16", len(msg.SessionKey))
			}

			// The LM response is the LMv2 response (HMAC || client challenge), never
			// the DES-based LM response derived from the password.
			if len(msg.LmChallengeResponse) != 24 {
				t.Fatalf("LmChallengeResponse is %d bytes, want 24", len(msg.LmChallengeResponse))
			}
			if bytes.Equal(msg.LmChallengeResponse, make([]byte, 24)) {
				t.Error("LmChallengeResponse is zeroed, want the LMv2 response")
			}

			sessionBaseKey := sessionBaseKeyOf(t, msg.NtChallengeResponse)
			keyExchange := tc.negFlags&flags.NTLMSSP_NEGOTIATE_KEY_EXCH != 0
			if got := msg.NegotiateFlags&flags.NTLMSSP_NEGOTIATE_KEY_EXCH != 0; got != keyExchange {
				t.Fatalf("KEY_EXCH asserted = %v, want %v", got, keyExchange)
			}
			if keyExchange {
				if len(msg.EncryptedRandomSessionKey) != 16 {
					t.Fatalf("EncryptedRandomSessionKey is %d bytes, want 16 with KEY_EXCH", len(msg.EncryptedRandomSessionKey))
				}
				c, err := rc4.NewCipher(sessionBaseKey)
				if err != nil {
					t.Fatalf("rc4.NewCipher: %v", err)
				}
				unwrapped := make([]byte, 16)
				c.XORKeyStream(unwrapped, msg.EncryptedRandomSessionKey)
				if !bytes.Equal(unwrapped, msg.SessionKey) {
					t.Error("EncryptedRandomSessionKey does not unwrap to SessionKey under the SessionBaseKey")
				}
			} else {
				if len(msg.EncryptedRandomSessionKey) != 0 {
					t.Errorf("EncryptedRandomSessionKey is %d bytes, want empty without KEY_EXCH", len(msg.EncryptedRandomSessionKey))
				}
				if !bytes.Equal(msg.SessionKey, sessionBaseKey) {
					t.Error("SessionKey is not the SessionBaseKey without KEY_EXCH")
				}
			}
		})
	}
}
