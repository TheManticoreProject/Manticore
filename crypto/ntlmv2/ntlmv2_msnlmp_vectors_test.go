package ntlmv2

import (
	"encoding/hex"
	"testing"
)

// TestMSNLMP_4_2_4_Vectors verifies the NTLMv2 computation against the official
// worked example in [MS-NLMP] section 4.2.4 (User / Domain / Password). It guards
// the NTOWFv2 identity construction in particular: only the username is
// uppercased, the domain is used as-is.
func TestMSNLMP_4_2_4_Vectors(t *testing.T) {
	mustHex := func(s string) []byte {
		b, err := hex.DecodeString(s)
		if err != nil {
			t.Fatalf("bad hex %q: %v", s, err)
		}
		return b
	}

	var serverChallenge, clientChallenge [8]byte
	copy(serverChallenge[:], mustHex("0123456789abcdef"))
	copy(clientChallenge[:], mustHex("aaaaaaaaaaaaaaaa"))

	ctx, err := NewNTLMv2CtxWithPassword("Domain", "User", "Password", serverChallenge, clientChallenge)
	if err != nil {
		t.Fatalf("NewNTLMv2CtxWithPassword: %v", err)
	}

	// MS-NLMP 4.2.4.1.1: NTOWFv2.
	if got := hex.EncodeToString(ctx.ResponseKeyNT[:]); got != "0c868a403bfd7a93a3001ef22ef02e3f" {
		t.Errorf("NTOWFv2 = %s, want 0c868a403bfd7a93a3001ef22ef02e3f", got)
	}

	// MS-NLMP 4.2.4.1.3 TargetInfo: NbDomainName "Domain", NbComputerName "Server", EOL.
	targetInfo := mustHex("02000c0044006f006d00610069006e0001000c0053006500720076006500720000000000")
	// Time is all-zero in the worked example.
	_, ntProofStr, err := ctx.ComputeNTChallengeResponse(make([]byte, 8), targetInfo)
	if err != nil {
		t.Fatalf("ComputeNTChallengeResponse: %v", err)
	}
	if got := hex.EncodeToString(ntProofStr); got != "68cd0ab851e51c96aabc927bebef6a1c" {
		t.Errorf("NTProofStr = %s, want 68cd0ab851e51c96aabc927bebef6a1c", got)
	}

	if got := hex.EncodeToString(ctx.ComputeSessionBaseKey(ntProofStr)); got != "8de40ccadbc14a82f15cb0ad0de95ca3" {
		t.Errorf("SessionBaseKey = %s, want 8de40ccadbc14a82f15cb0ad0de95ca3", got)
	}
}
