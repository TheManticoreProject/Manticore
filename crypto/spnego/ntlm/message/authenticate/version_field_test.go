package authenticate_test

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
)

// verChallenge builds a CHALLENGE with a server TargetInfo carrying an
// MsvAvTimestamp, as a Windows domain controller sends.
func verChallenge(t *testing.T) *challenge.ChallengeMessage {
	t.Helper()

	ts := make([]byte, 8)
	binary.LittleEndian.PutUint64(ts, (uint64(time.Now().Unix())+116444736000)*10000000)

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

// TestAuthenticateCarriesVersion guards the defect: the Version was computed and
// then discarded, because the assignment was gated on the server's advertised flags
// while Marshal gates on the client's own. [MS-NLMP] 2.2.1.3 makes Version a fixed
// field, so dropping it shifts every payload offset 8 bytes down from what a server
// computes.
func TestAuthenticateCarriesVersion(t *testing.T) {
	msg, err := authenticate.CreateAuthenticateMessage(
		verChallenge(t), "Administrator", "pass", "TMP", "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}

	if !msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_VERSION) {
		t.Error("AUTHENTICATE does not assert NTLMSSP_NEGOTIATE_VERSION")
	}
	if msg.Version == nil {
		t.Fatal("AUTHENTICATE carries no Version")
	}
	if want := version.DefaultVersion(); *msg.Version != want {
		t.Errorf("Version = %+v, want %+v", *msg.Version, want)
	}
	if msg.Version.ProductBuild == 0 {
		t.Error("Version.ProductBuild is 0, which is not a real Windows build")
	}

	// The payload must begin after the 64-byte fixed part plus the 8-byte Version,
	// which is where a server reads it from.
	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) < 64 {
		t.Fatalf("marshalled AUTHENTICATE is %d bytes", len(raw))
	}
	// DomainNameFields.BufferOffset sits at offset 32 of the fixed part and is the
	// first payload field in the canonical layout this marshaller emits.
	if got := binary.LittleEndian.Uint32(raw[32:36]); got != 72 {
		t.Errorf("first payload field begins at offset %d, want 72 (64-byte fixed part plus the 8-byte Version)", got)
	}
}
