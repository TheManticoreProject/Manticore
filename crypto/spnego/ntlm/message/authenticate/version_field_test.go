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

	// The payload begins after the 64-byte fixed part, the 8-byte Version and,
	// when one is carried, the 16-byte MIC ([MS-NLMP] 2.2.1.3). This challenge
	// supplies an MsvAvTimestamp, so a MIC is emitted and the payload starts at 88.
	assertFirstPayloadOffset(t, msg, 88)
}

// TestAuthenticateVersionOffsetWithoutMIC covers the other layout: a server that
// supplies no MsvAvTimestamp gets no MIC, and the payload then starts at 72.
func TestAuthenticateVersionOffsetWithoutMIC(t *testing.T) {
	ch := verChallenge(t)
	// Rebuild the TargetInfo without a timestamp, which is what suppresses the MIC.
	info, err := targetinfo.BuildServerTargetInfo("DC01", "TMP", "dc01.tmp.local", "tmp.local", nil)
	if err != nil {
		t.Fatalf("BuildServerTargetInfo: %v", err)
	}
	ch.TargetInfo = info

	msg, err := authenticate.CreateAuthenticateMessage(ch, "Administrator", "pass", "TMP", "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}
	if msg.NeedsMIC {
		t.Fatal("a challenge without MsvAvTimestamp produced a MIC")
	}
	assertFirstPayloadOffset(t, msg, 72)
}

// assertFirstPayloadOffset marshals the message and checks where the payload
// begins, which is where a server reads it from.
func assertFirstPayloadOffset(t *testing.T, msg *authenticate.AuthenticateMessage, want uint32) {
	t.Helper()

	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) < 64 {
		t.Fatalf("marshalled AUTHENTICATE is %d bytes", len(raw))
	}
	// DomainNameFields.BufferOffset sits at offset 32 of the fixed part and is the
	// first payload field in the canonical layout this marshaller emits.
	if got := binary.LittleEndian.Uint32(raw[32:36]); got != want {
		t.Errorf("first payload field begins at offset %d, want %d", got, want)
	}
}
