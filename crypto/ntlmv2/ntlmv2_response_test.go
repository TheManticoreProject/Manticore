package ntlmv2

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

// mustDecodeHex decodes a hex literal in a test, failing rather than returning.
func mustDecodeHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}

// The published hashcat example hash for mode 5600, split into its fields. It is
// the authoritative statement of the layout, so reproducing it exactly is the
// strongest available check that the renderer is right.
//
// Source: https://hashcat.net/wiki/doku.php?id=example_hashes
const (
	exampleUsername        = "admin"
	exampleDomain          = "N46iSNekpT"
	exampleServerChallenge = "08ca45b7d7ea58ee"
	exampleNTProofStr      = "88dcbe4446168966a153a0064958dac6"
	exampleBlob            = "5c7830315c7830310000000000000b45c67103d07d7b95acd12ffa11230e0000000052920b85f78d013c31cdb3b92f5d765c783030"
	exampleHash            = exampleUsername + "::" + exampleDomain + ":" + exampleServerChallenge + ":" + exampleNTProofStr + ":" + exampleBlob
)

// TestHashcatStringMatchesPublishedExample asserts the renderer reproduces the
// published mode-5600 example hash byte for byte, which pins both the field
// order and the field contents.
//
// The previous implementation emitted the server challenge followed by the LM and
// NT responses, which is neither the mode-5600 layout nor the mode-5500 one, so
// nothing it produced for an NTLMv2 capture was crackable.
func TestHashcatStringMatchesPublishedExample(t *testing.T) {
	var serverChallenge [8]byte
	copy(serverChallenge[:], mustDecodeHex(t, exampleServerChallenge))

	ntChallengeResponse := append(mustDecodeHex(t, exampleNTProofStr), mustDecodeHex(t, exampleBlob)...)

	response := NewNTLMv2Response(exampleUsername, exampleDomain, serverChallenge, [24]byte{}, ntChallengeResponse)

	got, err := response.HashcatString()
	if err != nil {
		t.Fatalf("HashcatString() error = %v", err)
	}
	// The renderer uppercases hex; hashcat parses it case-insensitively, so the
	// comparison is too.
	if !strings.EqualFold(got, exampleHash) {
		t.Fatalf("HashcatString() =\n  %s\nwant (any case)\n  %s", got, exampleHash)
	}
}

// TestHashcatStringFieldLayout asserts the rendered line has exactly the fields
// mode 5600 defines, in order, with the lengths each one is fixed at.
func TestHashcatStringFieldLayout(t *testing.T) {
	var serverChallenge [8]byte
	copy(serverChallenge[:], mustDecodeHex(t, exampleServerChallenge))
	ntChallengeResponse := append(mustDecodeHex(t, exampleNTProofStr), mustDecodeHex(t, exampleBlob)...)

	response := NewNTLMv2Response("user", "DOMAIN", serverChallenge, [24]byte{}, ntChallengeResponse)
	line, err := response.HashcatString()
	if err != nil {
		t.Fatalf("HashcatString() error = %v", err)
	}

	fields := strings.Split(line, ":")
	if len(fields) != 6 {
		t.Fatalf("line has %d colon-separated fields, want 6 (the second is empty): %q", len(fields), line)
	}
	if fields[0] != "user" {
		t.Fatalf("field 1 (username) = %q", fields[0])
	}
	if fields[1] != "" {
		t.Fatalf("field 2 should be empty, got %q", fields[1])
	}
	if fields[2] != "DOMAIN" {
		t.Fatalf("field 3 (domain) = %q", fields[2])
	}
	if len(fields[3]) != 16 {
		t.Fatalf("field 4 (server challenge) is %d hex chars, want 16 (8 bytes): %q", len(fields[3]), fields[3])
	}
	if len(fields[4]) != NTProofStrLength*2 {
		t.Fatalf("field 5 (NTProofStr) is %d hex chars, want %d (%d bytes): %q",
			len(fields[4]), NTProofStrLength*2, NTProofStrLength, fields[4])
	}
	if len(fields[5]) != len(exampleBlob) {
		t.Fatalf("field 6 (blob) is %d hex chars, want %d", len(fields[5]), len(exampleBlob))
	}
}

// TestHashcatStringFromComputedResponse renders a response computed by this
// package from the official [MS-NLMP] 4.2.4 worked example, tying the renderer to
// the spec vector rather than only to a hand-assembled line.
func TestHashcatStringFromComputedResponse(t *testing.T) {
	var serverChallenge, clientChallenge [8]byte
	copy(serverChallenge[:], mustDecodeHex(t, "0123456789abcdef"))
	copy(clientChallenge[:], mustDecodeHex(t, "aaaaaaaaaaaaaaaa"))

	ctx, err := NewNTLMv2CtxWithPassword("Domain", "User", "Password", serverChallenge, clientChallenge)
	if err != nil {
		t.Fatalf("NewNTLMv2CtxWithPassword() error = %v", err)
	}

	// [MS-NLMP] 4.2.4.1.3 ServerName, as used by the existing vectors test.
	targetInfo := mustDecodeHex(t, "02000c0044006f006d00610069006e0001000c005300650072007600650072000000000000000000")
	ntChallengeResponse, ntProofStr, err := ctx.ComputeNTChallengeResponse(make([]byte, 8), targetInfo)
	if err != nil {
		t.Fatalf("ComputeNTChallengeResponse() error = %v", err)
	}

	response := NewNTLMv2Response("User", "Domain", serverChallenge, [24]byte{}, ntChallengeResponse)

	// The NTProofStr the accessor reports must be the one the computation
	// produced, and the one [MS-NLMP] 4.2.4.2.2 publishes.
	if !bytes.Equal(response.NTProofStr(), ntProofStr) {
		t.Fatalf("NTProofStr() = %x, want %x", response.NTProofStr(), ntProofStr)
	}
	if got := hex.EncodeToString(response.NTProofStr()); got != "68cd0ab851e51c96aabc927bebef6a1c" {
		t.Fatalf("NTProofStr() = %s, want the MS-NLMP 4.2.4 value 68cd0ab851e51c96aabc927bebef6a1c", got)
	}

	line, err := response.HashcatString()
	if err != nil {
		t.Fatalf("HashcatString() error = %v", err)
	}
	fields := strings.Split(line, ":")
	if len(fields) != 6 {
		t.Fatalf("line has %d fields, want 6: %q", len(fields), line)
	}
	if !strings.EqualFold(fields[3], "0123456789abcdef") {
		t.Fatalf("server challenge field = %q, want the vector's 0123456789abcdef", fields[3])
	}
	if !strings.EqualFold(fields[4], "68cd0ab851e51c96aabc927bebef6a1c") {
		t.Fatalf("NTProofStr field = %q, want 68cd0ab851e51c96aabc927bebef6a1c", fields[4])
	}
	// Reassembling fields 5 and 6 must give back the whole NT response.
	reassembled := mustDecodeHex(t, fields[4]+fields[5])
	if !bytes.Equal(reassembled, ntChallengeResponse) {
		t.Fatalf("NTProofStr||blob does not reassemble to the NT challenge response")
	}
}

// TestVariableLengthResponseIsPreserved asserts a realistic NT challenge response
// is held whole. The field was a fixed [24]byte, which could not represent one at
// all: a real response is the 16-byte NTProofStr plus a blob that is itself at
// least 28 bytes before the TargetInfo is appended.
func TestVariableLengthResponseIsPreserved(t *testing.T) {
	ntChallengeResponse := append(mustDecodeHex(t, exampleNTProofStr), mustDecodeHex(t, exampleBlob)...)
	if len(ntChallengeResponse) <= 24 {
		t.Fatalf("the test vector is only %d bytes, so it does not exercise the variable length", len(ntChallengeResponse))
	}

	response := NewNTLMv2Response("user", "DOMAIN", [8]byte{}, [24]byte{}, ntChallengeResponse)

	if len(response.NtChallengeResponse) != len(ntChallengeResponse) {
		t.Fatalf("stored response is %d bytes, want %d", len(response.NtChallengeResponse), len(ntChallengeResponse))
	}
	if !bytes.Equal(response.NtChallengeResponse, ntChallengeResponse) {
		t.Fatal("stored response does not match what was supplied")
	}
	if got := len(response.NTProofStr()); got != NTProofStrLength {
		t.Fatalf("NTProofStr() is %d bytes, want %d", got, NTProofStrLength)
	}
	if got, want := len(response.Blob()), len(ntChallengeResponse)-NTProofStrLength; got != want {
		t.Fatalf("Blob() is %d bytes, want %d", got, want)
	}
	if !bytes.Equal(response.Blob(), mustDecodeHex(t, exampleBlob)) {
		t.Fatal("Blob() does not match the blob that was supplied")
	}
}

// TestConstructorCopiesResponse asserts the constructor does not alias the
// caller's slice, so a caller reusing a receive buffer cannot silently rewrite a
// response that has already been recorded.
func TestConstructorCopiesResponse(t *testing.T) {
	ntChallengeResponse := append(mustDecodeHex(t, exampleNTProofStr), mustDecodeHex(t, exampleBlob)...)
	original := append([]byte(nil), ntChallengeResponse...)

	response := NewNTLMv2Response("user", "DOMAIN", [8]byte{}, [24]byte{}, ntChallengeResponse)

	// Scribble over the caller's buffer.
	for i := range ntChallengeResponse {
		ntChallengeResponse[i] = 0xAA
	}

	if !bytes.Equal(response.NtChallengeResponse, original) {
		t.Fatal("mutating the caller's slice changed the stored response")
	}
}

// TestHashcatStringRejectsUnrenderableResponses asserts a response that cannot
// form a mode-5600 line is reported as an error rather than producing a malformed
// line or slicing past the end of the buffer.
func TestHashcatStringRejectsUnrenderableResponses(t *testing.T) {
	cases := []struct {
		name                string
		ntChallengeResponse []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
		{"one byte short of the NTProofStr", make([]byte, NTProofStrLength-1)},
		{"exactly the NTProofStr, no blob", make([]byte, NTProofStrLength)},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			response := NewNTLMv2Response("user", "DOMAIN", [8]byte{}, [24]byte{}, tc.ntChallengeResponse)

			line, err := response.HashcatString()
			if err == nil {
				t.Fatalf("HashcatString() returned %q, want an error", line)
			}
			if line != "" {
				t.Fatalf("HashcatString() returned %q alongside an error, want the empty string", line)
			}
			if got := response.String(); got != "" {
				t.Fatalf("String() = %q, want the empty string when the line cannot be rendered", got)
			}
		})
	}
}

// TestAccessorsOnShortResponse asserts the accessors report absence rather than
// panicking on a response too short to carry what is asked for.
func TestAccessorsOnShortResponse(t *testing.T) {
	short := NewNTLMv2Response("user", "DOMAIN", [8]byte{}, [24]byte{}, make([]byte, NTProofStrLength-1))
	if got := short.NTProofStr(); got != nil {
		t.Fatalf("NTProofStr() = %x on a short response, want nil", got)
	}
	if got := short.Blob(); got != nil {
		t.Fatalf("Blob() = %x on a short response, want nil", got)
	}

	exact := NewNTLMv2Response("user", "DOMAIN", [8]byte{}, [24]byte{}, make([]byte, NTProofStrLength))
	if got := exact.NTProofStr(); len(got) != NTProofStrLength {
		t.Fatalf("NTProofStr() is %d bytes on an exact-length response, want %d", len(got), NTProofStrLength)
	}
	if got := exact.Blob(); got != nil {
		t.Fatalf("Blob() = %x when there is nothing after the NTProofStr, want nil", got)
	}
}

// TestStringRendersTheHashcatLine asserts String is the hashcat line when one can
// be rendered, which is what makes the type printable straight into an output
// file.
func TestStringRendersTheHashcatLine(t *testing.T) {
	var serverChallenge [8]byte
	copy(serverChallenge[:], mustDecodeHex(t, exampleServerChallenge))
	ntChallengeResponse := append(mustDecodeHex(t, exampleNTProofStr), mustDecodeHex(t, exampleBlob)...)

	response := NewNTLMv2Response(exampleUsername, exampleDomain, serverChallenge, [24]byte{}, ntChallengeResponse)

	line, err := response.HashcatString()
	if err != nil {
		t.Fatalf("HashcatString() error = %v", err)
	}
	if got := response.String(); got != line {
		t.Fatalf("String() = %q, want the hashcat line %q", got, line)
	}
}

// TestHashcatModeIsNetNTLMv2 guards the mode number a caller labels captured
// material with. A NetNTLMv2 response written under mode 5500 is unusable.
func TestHashcatModeIsNetNTLMv2(t *testing.T) {
	if HashcatMode != 5600 {
		t.Fatalf("HashcatMode = %d, want 5600 for NetNTLMv2", HashcatMode)
	}
}

// TestLengthConstants guards the two fixed lengths the layout depends on.
func TestLengthConstants(t *testing.T) {
	if NTProofStrLength != 16 {
		t.Fatalf("NTProofStrLength = %d, want 16", NTProofStrLength)
	}
	if LmChallengeResponseLength != 24 {
		t.Fatalf("LmChallengeResponseLength = %d, want 24", LmChallengeResponseLength)
	}
}
