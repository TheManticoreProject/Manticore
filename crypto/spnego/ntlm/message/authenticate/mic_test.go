package authenticate_test

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
)

// micChallenge builds a CHALLENGE with a server TargetInfo, carrying an
// MsvAvTimestamp when withTimestamp is set — the condition under which a Windows
// client emits a MIC.
func micChallenge(t *testing.T, withTimestamp bool) *challenge.ChallengeMessage {
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

// avPairValue walks a TargetInfo and returns the value of the first pair with id.
func avPairValue(t *testing.T, info []byte, id avpair.AvId) ([]byte, bool) {
	t.Helper()

	for i := 0; i+4 <= len(info); {
		current := avpair.AvId(binary.LittleEndian.Uint16(info[i : i+2]))
		avLen := int(binary.LittleEndian.Uint16(info[i+2 : i+4]))
		if i+4+avLen > len(info) {
			return nil, false
		}
		if current == id {
			return info[i+4 : i+4+avLen], true
		}
		if current == avpair.MsvAvEOL {
			return nil, false
		}
		i += 4 + avLen
	}
	return nil, false
}

// TestAuthenticateSignalsAndCarriesAMIC guards the defect: NeedsMIC was hardcoded
// false, so the MsvAvFlags pair was never added and the MIC never computed. A
// server detects a MIC solely by MsvAvFlags bit 0x2 ([MS-NLMP] 3.2.5.1.2), so the
// flag and the field have to agree or the MIC is ignored.
func TestAuthenticateSignalsAndCarriesAMIC(t *testing.T) {
	msg, err := authenticate.CreateAuthenticateMessage(
		micChallenge(t, true), "Administrator", "pass", "TMP", "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}

	if !msg.NeedsMIC {
		t.Fatal("a CHALLENGE carrying MsvAvTimestamp produced no MIC")
	}

	// The NTLMv2 response is NTProofStr(16) followed by the blob, whose TargetInfo
	// must now carry MsvAvFlags with the MIC-present bit.
	if len(msg.NtChallengeResponse) <= 16 {
		t.Fatalf("NtChallengeResponse is %d bytes, too short to hold a blob", len(msg.NtChallengeResponse))
	}
	blob := msg.NtChallengeResponse[16:]
	// The blob header is 28 bytes before its TargetInfo (MS-NLMP 2.2.2.7).
	if len(blob) <= 28 {
		t.Fatalf("NTLMv2 blob is %d bytes, too short to hold a TargetInfo", len(blob))
	}
	value, ok := avPairValue(t, blob[28:], avpair.MsvAvFlags)
	if !ok {
		t.Fatal("the NTLMv2 blob carries no MsvAvFlags pair, so a server will not look for a MIC")
	}
	if len(value) != 4 {
		t.Fatalf("MsvAvFlags value is %d bytes, want 4", len(value))
	}
	if got := binary.LittleEndian.Uint32(value); got&avpair.MsvAvFlagMICPresent == 0 {
		t.Errorf("MsvAvFlags = %#08x, missing the MIC-present bit %#08x", got, avpair.MsvAvFlagMICPresent)
	}

	// The MIC itself is computed by the caller that holds the NEGOTIATE bytes, so
	// here it is still zero; what must be true now is that Marshal reserves it.
	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := binary.LittleEndian.Uint32(raw[32:36]); got != 88 {
		t.Errorf("first payload field begins at offset %d, want 88 (64 fixed + 8 Version + 16 MIC)", got)
	}
}

// TestAuthenticateWithoutTimestampCarriesNoMIC pins the other branch: a server
// that supplies no MsvAvTimestamp gets no MsvAvFlags pair and no reserved MIC.
func TestAuthenticateWithoutTimestampCarriesNoMIC(t *testing.T) {
	msg, err := authenticate.CreateAuthenticateMessage(
		micChallenge(t, false), "Administrator", "pass", "TMP", "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}

	if msg.NeedsMIC {
		t.Error("a CHALLENGE without MsvAvTimestamp produced a MIC")
	}
	blob := msg.NtChallengeResponse[16:]
	if _, ok := avPairValue(t, blob[28:], avpair.MsvAvFlags); ok {
		t.Error("MsvAvFlags was added although no MIC will be carried")
	}
}

// TestComputeMICChangesTheMessage checks the MIC is actually written and depends
// on the handshake messages it covers.
func TestComputeMICChangesTheMessage(t *testing.T) {
	msg, err := authenticate.CreateAuthenticateMessage(
		micChallenge(t, true), "Administrator", "pass", "TMP", "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}

	negotiate := []byte("NTLMSSP\x00\x01\x00\x00\x00pretend-negotiate")
	challengeBytes := []byte("NTLMSSP\x00\x02\x00\x00\x00pretend-challenge")

	if err := msg.ComputeMIC(negotiate, challengeBytes); err != nil {
		t.Fatalf("ComputeMIC: %v", err)
	}
	first := msg.MIC
	if first == ([16]byte{}) {
		t.Fatal("ComputeMIC left the MIC zeroed")
	}

	// A different NEGOTIATE must produce a different MIC, or it binds nothing.
	if err := msg.ComputeMIC([]byte("NTLMSSP\x00\x01\x00\x00\x00different"), challengeBytes); err != nil {
		t.Fatalf("ComputeMIC (second): %v", err)
	}
	if bytes.Equal(first[:], msg.MIC[:]) {
		t.Error("the MIC is unchanged by a different NEGOTIATE message, so it binds nothing")
	}
}
