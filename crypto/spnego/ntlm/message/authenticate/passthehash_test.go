package authenticate

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
)

// newTestChallenge builds a minimal CHALLENGE_MESSAGE for driving AUTHENTICATE
// construction in tests, with the given negotiate flags and a fixed server challenge.
func newTestChallenge(negFlags flags.NegotiateFlags) *challenge.ChallengeMessage {
	return &challenge.ChallengeMessage{
		NegotiateFlags:  negFlags,
		ServerChallenge: [8]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef},
	}
}

// TestCreateAuthenticateMessageWithNTHash verifies the pass-the-hash entry point
// produces a well-formed NTLMv2 AUTHENTICATE message: a real NT/LM response and a
// 16-byte exported session key (needed for RPC/SMB signing and sealing).
func TestCreateAuthenticateMessageWithNTHash(t *testing.T) {
	negFlags := flags.NTLMSSP_NEGOTIATE_UNICODE |
		flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
		flags.NTLMSSP_NEGOTIATE_TARGET_INFO
	chal := newTestChallenge(negFlags)
	ntHash := nt.NTHash("Password")

	msg, err := CreateAuthenticateMessageWithNTHash(chal, "User", ntHash, "Domain", "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessageWithNTHash: %v", err)
	}
	if len(msg.SessionKey) != 16 {
		t.Errorf("SessionKey length = %d, want 16", len(msg.SessionKey))
	}
	if len(msg.NtChallengeResponse) < 16 {
		t.Errorf("NtChallengeResponse too short: %d bytes", len(msg.NtChallengeResponse))
	}
	if len(msg.LmChallengeResponse) != 24 {
		t.Errorf("LmChallengeResponse length = %d, want 24", len(msg.LmChallengeResponse))
	}

	// The message must marshal and round-trip back to a parseable AUTHENTICATE.
	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got AuthenticateMessage
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
}

// TestCreateAuthenticateMessageWithNTHashRequiresNTLMv2 confirms pass-the-hash is
// rejected when the server did not negotiate extended session security, since no
// session key could be derived for signing/sealing.
func TestCreateAuthenticateMessageWithNTHashRequiresNTLMv2(t *testing.T) {
	chal := newTestChallenge(flags.NTLMSSP_NEGOTIATE_UNICODE)
	if _, err := CreateAuthenticateMessageWithNTHash(chal, "User", nt.NTHash("Password"), "Domain", "WORKSTATION"); err == nil {
		t.Fatal("expected error without extended session security, got nil")
	}
}
