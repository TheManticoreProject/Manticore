package ntlmv2_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/ntlmv2"
)

// MS-NLMP spec test vectors — Appendix B / Section 4.2.4
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp

// TestNTOWFv2 verifies that NTOWFv2 (ResponseKeyNT derivation) follows the MS-NLMP spec:
//   ResponseKeyNT = HMAC-MD5(MD4(UNICODE(Passwd)), UNICODE(ConcatenationOf(Uppercase(User), UserDom)))
//
// With User="User", UserDom="Domain", Password="Password":
//   NT Hash = a4f49c406510bdcab6824ee7c30fd852
//   input   = UTF16-LE("USER" + "Domain") = UTF16-LE("USERDomain")
//   result  = HMAC-MD5(NT_Hash, input) = 0c868a403bfd7a93a3001ef22ef02e3f
func TestNTOWFv2(t *testing.T) {
	expected := "0c868a403bfd7a93a3001ef22ef02e3f"
	got := hex.EncodeToString(ntlmv2.NTOWFv2("Password", "User", "Domain"))
	if got != expected {
		t.Errorf("NTOWFv2: expected %s, got %s", expected, got)
	}
}

// TestLMOWFv2 verifies that LMOWFv2 equals NTOWFv2 (they are identical for NTLMv2).
func TestLMOWFv2(t *testing.T) {
	lm := hex.EncodeToString(ntlmv2.LMOWFv2("Password", "User", "Domain"))
	nt := hex.EncodeToString(ntlmv2.NTOWFv2("Password", "User", "Domain"))
	if lm != nt {
		t.Errorf("LMOWFv2 should equal NTOWFv2: lm=%s nt=%s", lm, nt)
	}
}

// TestComputeLMChallengeResponse_WithTimestamp verifies Z(24) is returned when
// MsvAvTimestamp is present, as required by MS-NLMP 3.1.5.1.2.
func TestComputeLMChallengeResponse_WithTimestamp(t *testing.T) {
	sc := [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	cc := [8]byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x11, 0x22}
	ctx, err := ntlmv2.NewNTLMv2CtxWithPassword("Domain", "User", "Password", sc, cc)
	if err != nil {
		t.Fatal(err)
	}
	lm := ctx.ComputeLMChallengeResponse(true)
	if len(lm) != 24 {
		t.Errorf("expected 24 bytes, got %d", len(lm))
	}
	if !bytes.Equal(lm, make([]byte, 24)) {
		t.Error("expected Z(24) when hasTimestamp=true")
	}
}

// TestComputeLMChallengeResponse_WithoutTimestamp verifies a 24-byte response
// (16-byte HMAC + 8-byte client challenge) when no MsvAvTimestamp.
func TestComputeLMChallengeResponse_WithoutTimestamp(t *testing.T) {
	sc := [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	cc := [8]byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x11, 0x22}
	ctx, err := ntlmv2.NewNTLMv2CtxWithPassword("Domain", "User", "Password", sc, cc)
	if err != nil {
		t.Fatal(err)
	}
	lm := ctx.ComputeLMChallengeResponse(false)
	if len(lm) != 24 {
		t.Errorf("expected 24 bytes, got %d", len(lm))
	}
	// Must not be all zeros
	if bytes.Equal(lm, make([]byte, 24)) {
		t.Error("LmChallengeResponse should not be Z(24) when hasTimestamp=false")
	}
	// Last 8 bytes must be the client challenge
	if !bytes.Equal(lm[16:], cc[:]) {
		t.Errorf("last 8 bytes should be client challenge: expected %x, got %x", cc[:], lm[16:])
	}
}

// TestComputeNTChallengeResponse verifies the output structure and length.
func TestComputeNTChallengeResponse(t *testing.T) {
	sc := [8]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}
	cc := [8]byte{0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA}
	ctx, err := ntlmv2.NewNTLMv2CtxWithPassword("Domain", "User", "Password", sc, cc)
	if err != nil {
		t.Fatal(err)
	}

	timestamp := make([]byte, 8)                 // Z(8) for simplicity
	targetInfo := []byte{0x00, 0x00, 0x00, 0x00} // minimal: just EOL

	ntCR, ntProofStr, err := ctx.ComputeNTChallengeResponse(timestamp, targetInfo)
	if err != nil {
		t.Fatal(err)
	}

	// NTProofStr is 16 bytes
	if len(ntProofStr) != 16 {
		t.Errorf("NTProofStr should be 16 bytes, got %d", len(ntProofStr))
	}

	// NTChallengeResponse = NTProofStr(16) || blob(variable)
	// blob = RespType(1)+HiRespType(1)+Z(2)+Z(4)+timestamp(8)+CC(8)+Z(4)+targetInfo
	// (the TargetInfo is already MsvAvEOL-terminated; no extra trailing Z(4) is added).
	expectedBlobLen := 1 + 1 + 2 + 4 + 8 + 8 + 4 + len(targetInfo)
	if len(ntCR) != 16+expectedBlobLen {
		t.Errorf("NTChallengeResponse length: expected %d, got %d", 16+expectedBlobLen, len(ntCR))
	}

	// First 16 bytes of NTChallengeResponse must equal NTProofStr
	if !bytes.Equal(ntCR[:16], ntProofStr) {
		t.Error("NTChallengeResponse must start with NTProofStr")
	}

	// Blob layout: RespType(1)+HiRespType(1)+Reserved1(2)+Reserved2(4)+Timestamp(8)+CC(8)
	// CC starts at blob offset 16, so NTChallengeResponse offset 16+16=32
	blobCC := ntCR[16+16 : 16+24]
	if !bytes.Equal(blobCC, cc[:]) {
		t.Errorf("client challenge in blob: expected %x, got %x", cc[:], blobCC)
	}
}

// TestComputeNTChallengeResponse_InvalidTimestamp verifies error on wrong timestamp length.
func TestComputeNTChallengeResponse_InvalidTimestamp(t *testing.T) {
	sc := [8]byte{}
	cc := [8]byte{}
	ctx, err := ntlmv2.NewNTLMv2CtxWithPassword("Domain", "User", "Password", sc, cc)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = ctx.ComputeNTChallengeResponse([]byte{0x01, 0x02}, nil) // wrong length
	if err == nil {
		t.Error("expected error for invalid timestamp length")
	}
}

// TestComputeSessionBaseKey verifies that different NTProofStr values yield different keys.
func TestComputeSessionBaseKey(t *testing.T) {
	sc := [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	cc := [8]byte{0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA}
	ctx, _ := ntlmv2.NewNTLMv2CtxWithPassword("Domain", "User", "Password", sc, cc)

	ntProofStr1 := make([]byte, 16)
	ntProofStr2 := make([]byte, 16)
	ntProofStr2[0] = 0xFF

	key1 := ctx.ComputeSessionBaseKey(ntProofStr1)
	key2 := ctx.ComputeSessionBaseKey(ntProofStr2)

	if len(key1) != 16 {
		t.Errorf("expected 16-byte key, got %d", len(key1))
	}
	if bytes.Equal(key1, key2) {
		t.Error("different NTProofStr should yield different SessionBaseKey")
	}
}

// TestComputeResponse_MSNLMPSection4_2_4 is a known-answer test pinning the
// low-level ComputeResponse primitive to the MS-NLMP §4.2.4 worked example
// ("NTLMv2 Authentication"). Unlike the structural tests above, it compares
// against the *published NTProofStr and LM MAC values* so that any regression
// in the HMAC construction, blob layout, or LM response format will fail here
// before reaching a live authentication attempt.
//
// Reference: MS-NLMP §4.2.4
func TestComputeResponse_MSNLMPSection4_2_4(t *testing.T) {
	// NTOWFv2(Password="Password", User="User", UserDomain="Domain")
	responseKeyNT, _ := hex.DecodeString("0c868a403bfd7a93a3001ef22ef02e3f")
	responseKeyLM := responseKeyNT

	serverChallenge, _ := hex.DecodeString("0123456789abcdef")
	clientChallenge, _ := hex.DecodeString("aaaaaaaaaaaaaaaa")

	// TargetInfo from §4.2.4.1.3
	serverName, _ := hex.DecodeString(
		"02000c0044006f006d00610069006e00" +
			"01000c00530065007200760065007200" +
			"00000000",
	)

	// The spec fixes the timestamp to all zeros.
	timestamp := make([]byte, 8)

	ctx := &ntlmv2.NTLMv2Ctx{}
	got, err := ctx.ComputeResponse(responseKeyNT, responseKeyLM, serverChallenge, clientChallenge, timestamp, serverName)
	if err != nil {
		t.Fatalf("ComputeResponse returned error: %v", err)
	}

	expected, _ := hex.DecodeString(
		// NTProofStr (16 bytes)
		"68cd0ab851e51c96aabc927bebef6a1c" +
			// temp blob
			"0101000000000000" + // RespType=1, HiRespType=1, Reserved=Z(6)
			"0000000000000000" + // Timestamp (FILETIME, LE, = 0)
			"aaaaaaaaaaaaaaaa" + // ClientChallenge
			"00000000" + // Reserved Z(4)
			"02000c0044006f006d00610069006e00" + // ServerName (TargetInfo)
			"01000c00530065007200760065007200" +
			"00000000" +
			"00000000" + // Z(4)
			// LmChallengeResponse = HMAC_MD5(...) || ClientChallenge
			"86c35097ac9cec102554764a57cccc19" +
			"aaaaaaaaaaaaaaaa",
	)

	if !bytes.Equal(got, expected) {
		t.Fatalf("ComputeResponse mismatch vs MS-NLMP §4.2.4\n got:  %s\n want: %s",
			hex.EncodeToString(got), hex.EncodeToString(expected))
	}
}
