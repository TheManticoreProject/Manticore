package authenticate_test

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/authenticate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
)

// filetimeEpochDelta is the number of 100 ns ticks between the Windows FILETIME
// epoch (1601-01-01) and the Unix epoch (1970-01-01).
const filetimeEpochDelta = 116444736000000000

// TestFallbackTimestampIsCurrentTime guards against a far-future NTLMv2 blob
// TimeStamp. When the CHALLENGE carries no MsvAvTimestamp the client derives its
// own FILETIME; a wrong epoch offset once put the value near the year 5348, a
// standalone fingerprint. The derived timestamp must decode to within a minute
// of the current time.
func TestFallbackTimestampIsCurrentTime(t *testing.T) {
	// A CHALLENGE with no TargetInfo (hence no MsvAvTimestamp) forces the client
	// onto the self-derived FILETIME path.
	ch := noESSChallenge(flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_NTLM)

	before := time.Now()
	msg, err := authenticate.CreateAuthenticateMessage(ch, kxUser, kxPassword, kxDomain, "WORKSTATION")
	if err != nil {
		t.Fatalf("CreateAuthenticateMessage: %v", err)
	}
	after := time.Now()

	// The NTLMv2 blob TimeStamp is the 8 bytes at offset 24 of the NT response:
	// 16-byte NTProofStr, then the blob header (RespType, HiRespType, 2 reserved,
	// 4 reserved) then the 8-byte TimeStamp.
	if len(msg.NtChallengeResponse) < 32 {
		t.Fatalf("NtChallengeResponse is %d bytes, too short to hold a TimeStamp", len(msg.NtChallengeResponse))
	}
	filetime := binary.LittleEndian.Uint64(msg.NtChallengeResponse[24:32])

	if filetime < filetimeEpochDelta {
		t.Fatalf("FILETIME %#x decodes before the Unix epoch, want a current timestamp", filetime)
	}
	unixSeconds := int64((filetime - filetimeEpochDelta) / 10000000)
	got := time.Unix(unixSeconds, 0).UTC()

	lo := before.Add(-time.Minute)
	hi := after.Add(time.Minute)
	if got.Before(lo) || got.After(hi) {
		t.Fatalf("fallback TimeStamp decodes to %s (FILETIME %#x), want ~now in [%s, %s]", got, filetime, lo.UTC(), hi.UTC())
	}
}
