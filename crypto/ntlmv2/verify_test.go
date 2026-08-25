package ntlmv2

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// msnlmpVector returns the [MS-NLMP] 4.2.4 worked example: the context, the full
// NT challenge response and its NTProofStr.
func msnlmpVector(t *testing.T) (*NTLMv2Ctx, []byte, []byte) {
	t.Helper()

	decode := func(s string) []byte {
		b, err := hex.DecodeString(s)
		if err != nil {
			t.Fatalf("bad hex %q: %v", s, err)
		}
		return b
	}

	var serverChallenge, clientChallenge [8]byte
	copy(serverChallenge[:], decode("0123456789abcdef"))
	copy(clientChallenge[:], decode("aaaaaaaaaaaaaaaa"))

	ctx, err := NewNTLMv2CtxWithPassword("Domain", "User", "Password", serverChallenge, clientChallenge)
	if err != nil {
		t.Fatalf("NewNTLMv2CtxWithPassword() error = %v", err)
	}

	targetInfo := decode("02000c0044006f006d00610069006e0001000c005300650072007600650072000000000000000000")
	ntResponse, ntProofStr, err := ctx.ComputeNTChallengeResponse(make([]byte, 8), targetInfo)
	if err != nil {
		t.Fatalf("ComputeNTChallengeResponse() error = %v", err)
	}

	return ctx, ntResponse, ntProofStr
}

// TestVerifyNTChallengeResponseAcceptsTheSpecVector asserts a response computed
// from the [MS-NLMP] 4.2.4 worked example verifies, tying the acceptor-side check
// to the same published vector the computing side is checked against.
func TestVerifyNTChallengeResponseAcceptsTheSpecVector(t *testing.T) {
	ctx, ntResponse, ntProofStr := msnlmpVector(t)

	// Guard the vector itself, so a failure below is about verification rather
	// than about the response having changed.
	if got := hex.EncodeToString(ntProofStr); got != "68cd0ab851e51c96aabc927bebef6a1c" {
		t.Fatalf("NTProofStr = %s, want the MS-NLMP 4.2.4 value 68cd0ab851e51c96aabc927bebef6a1c", got)
	}

	if !VerifyNTChallengeResponse(ctx.ResponseKeyNT[:], ctx.ServerChallenge, ntResponse) {
		t.Fatal("VerifyNTChallengeResponse() rejected the response computed from the spec vector")
	}
}

// TestVerifyNTChallengeResponseRejects asserts every way a response can be wrong
// is rejected. The verifier is what stands between an acceptor and an attacker who
// does not hold the credential, so a false accept is the worst outcome available.
func TestVerifyNTChallengeResponseRejects(t *testing.T) {
	ctx, ntResponse, _ := msnlmpVector(t)
	key := ctx.ResponseKeyNT[:]

	t.Run("wrong response key", func(t *testing.T) {
		wrong := append([]byte(nil), key...)
		wrong[0] ^= 0x01
		if VerifyNTChallengeResponse(wrong, ctx.ServerChallenge, ntResponse) {
			t.Fatal("accepted a response under the wrong response key")
		}
	})

	t.Run("wrong server challenge", func(t *testing.T) {
		other := ctx.ServerChallenge
		other[0] ^= 0x01
		if VerifyNTChallengeResponse(key, other, ntResponse) {
			t.Fatal("accepted a response against a different challenge, so a captured response could be replayed")
		}
	})

	t.Run("tampered NTProofStr", func(t *testing.T) {
		tampered := append([]byte(nil), ntResponse...)
		tampered[0] ^= 0x01
		if VerifyNTChallengeResponse(key, ctx.ServerChallenge, tampered) {
			t.Fatal("accepted a response with a modified NTProofStr")
		}
	})

	t.Run("tampered blob", func(t *testing.T) {
		// A modified blob must invalidate the proof: the blob carries the
		// timestamp and the TargetInfo, so accepting a changed one would let a
		// client rewrite what it committed to.
		tampered := append([]byte(nil), ntResponse...)
		tampered[len(tampered)-1] ^= 0x01
		if VerifyNTChallengeResponse(key, ctx.ServerChallenge, tampered) {
			t.Fatal("accepted a response with a modified blob")
		}
	})

	t.Run("degenerate inputs", func(t *testing.T) {
		cases := []struct {
			name       string
			key        []byte
			ntResponse []byte
		}{
			{"nil key", nil, ntResponse},
			{"empty key", []byte{}, ntResponse},
			{"nil response", key, nil},
			{"response with no blob", key, ntResponse[:NTProofStrLength]},
			{"response shorter than the NTProofStr", key, ntResponse[:NTProofStrLength-1]},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				if VerifyNTChallengeResponse(tc.key, ctx.ServerChallenge, tc.ntResponse) {
					t.Fatal("accepted a degenerate input")
				}
			})
		}
	})
}

// TestSessionBaseKeyFromResponse asserts the acceptor's route to the
// SessionBaseKey agrees with the computing side's. The two start from different
// places — one from a response as received, the other from a proof it computed —
// so a disagreement would leave the two ends of an exchange with different keys.
func TestSessionBaseKeyFromResponse(t *testing.T) {
	ctx, ntResponse, ntProofStr := msnlmpVector(t)

	fromResponse := SessionBaseKeyFromResponse(ctx.ResponseKeyNT[:], ntResponse)
	fromComputation := ctx.ComputeSessionBaseKey(ntProofStr)

	if !bytes.Equal(fromResponse, fromComputation) {
		t.Fatalf("SessionBaseKeyFromResponse = %x, ComputeSessionBaseKey = %x", fromResponse, fromComputation)
	}
	// The [MS-NLMP] 4.2.4.2.2 published SessionBaseKey.
	if got := hex.EncodeToString(fromResponse); got != "8de40ccadbc14a82f15cb0ad0de95ca3" {
		t.Fatalf("SessionBaseKey = %s, want the MS-NLMP 4.2.4 value 8de40ccadbc14a82f15cb0ad0de95ca3", got)
	}

	// Degenerate inputs report absence rather than a short or wrong key.
	if got := SessionBaseKeyFromResponse(nil, ntResponse); got != nil {
		t.Fatalf("SessionBaseKeyFromResponse() with no key = %x, want nil", got)
	}
	if got := SessionBaseKeyFromResponse(ctx.ResponseKeyNT[:], ntResponse[:NTProofStrLength-1]); got != nil {
		t.Fatalf("SessionBaseKeyFromResponse() with a short response = %x, want nil", got)
	}
}
