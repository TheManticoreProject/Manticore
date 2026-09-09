package authenticate_test

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
)

// challengeWithTargetInfo builds a CHALLENGE carrying a server TargetInfo, with an
// MsvAvTimestamp when withTimestamp is set — which is what a Windows domain
// controller always sends.
func challengeWithTargetInfo(t *testing.T, withTimestamp bool) *challenge.ChallengeMessage {
	t.Helper()

	var ts []byte
	if withTimestamp {
		ts = make([]byte, 8)
		binary.LittleEndian.PutUint64(ts, (uint64(time.Now().Unix())+116444736000)*10000000)
	}

	info, err := targetinfo.BuildServerTargetInfo("DC01", "TMP", "dc01.tmp.local", "tmp.local", ts)
	if err != nil {
		t.Fatalf("BuildServerTargetInfo: %v", err)
	}

	msg := &challenge.ChallengeMessage{
		NegotiateFlags: flags.NTLMSSP_NEGOTIATE_UNICODE |
			flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
			flags.NTLMSSP_NEGOTIATE_TARGET_INFO,
		TargetInfo: info,
	}
	copy(msg.ServerChallenge[:], []byte{1, 2, 3, 4, 5, 6, 7, 8})
	return msg
}

// TestLmChallengeResponseIsZeroedWhenServerSuppliesTimestamp guards the defect: the
// caller passed a hardcoded false, so the conforming branch of
// ComputeLMChallengeResponse was dead. MS-NLMP 3.1.5.1.2 requires Z(24) when the
// CHALLENGE TargetInfo carries MsvAvTimestamp, which a Windows DC always does.
func TestLmChallengeResponseIsZeroedWhenServerSuppliesTimestamp(t *testing.T) {
	authMsg, err := authenticate.CreateAuthenticateMessage(
		challengeWithTargetInfo(t, true), "Administrator", "pass", "TMP", "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}

	if len(authMsg.LmChallengeResponse) != 24 {
		t.Fatalf("LmChallengeResponse is %d bytes, want 24", len(authMsg.LmChallengeResponse))
	}
	if !bytes.Equal(authMsg.LmChallengeResponse, make([]byte, 24)) {
		t.Errorf("LmChallengeResponse = % x, want 24 zero bytes", authMsg.LmChallengeResponse)
	}
}

// TestLmChallengeResponseIsComputedWithoutTimestamp checks the other branch is intact:
// with no MsvAvTimestamp the field carries the computed LMv2 response.
func TestLmChallengeResponseIsComputedWithoutTimestamp(t *testing.T) {
	authMsg, err := authenticate.CreateAuthenticateMessage(
		challengeWithTargetInfo(t, false), "Administrator", "pass", "TMP", "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}

	if len(authMsg.LmChallengeResponse) != 24 {
		t.Fatalf("LmChallengeResponse is %d bytes, want 24", len(authMsg.LmChallengeResponse))
	}
	if bytes.Equal(authMsg.LmChallengeResponse, make([]byte, 24)) {
		t.Error("LmChallengeResponse is zeroed, want the computed LMv2 response when no MsvAvTimestamp was supplied")
	}
	// The trailing 8 bytes are the client challenge, so they must not be zero either.
	if bytes.Equal(authMsg.LmChallengeResponse[16:], make([]byte, 8)) {
		t.Error("LmChallengeResponse client-challenge tail is zeroed")
	}
}
