package ntlmv2

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// HashcatMode is the hashcat hash-mode number for a NetNTLMv2 response, so a
// caller writing captured material to a file can label it correctly.
const HashcatMode = 5600

// NTProofStrLength is the length in bytes of the NTProofStr that prefixes an
// NTLMv2 NtChallengeResponse. It is the HMAC-MD5 taken over the server challenge
// concatenated with the blob that follows it ([MS-NLMP] 3.3.2).
const NTProofStrLength = 16

// LmChallengeResponseLength is the length in bytes of an LMv2 response: a
// 16-byte HMAC-MD5 followed by the 8-byte client challenge. Unlike the NT
// response it is fixed-size, and it is all zeroes when the server's TargetInfo
// carried a timestamp.
const LmChallengeResponseLength = 24

// NTLMv2Response is a NetNTLMv2 authentication: the identity that was claimed,
// the server challenge it answered, and the responses it carried.
//
// NtChallengeResponse is variable-length by construction —
// NTProofStr(16) || blob(variable), exactly what
// NTLMv2Ctx.ComputeNTChallengeResponse produces — where the blob holds the
// response type, a timestamp, the client challenge and the server's TargetInfo
// AV_PAIR list. In practice it runs from roughly 44 bytes to a few hundred, so it
// cannot be held in a fixed-size array.
//
// An NtChallengeResponse of exactly 24 bytes is an NTLMv1 response rather than an
// NTLMv2 one, and belongs in crypto/ntlmv1.NTLMv1Response, which renders the
// different layout hashcat expects for it.
type NTLMv2Response struct {
	Username string // Username for authentication
	Domain   string // Domain name

	ServerChallenge     [8]byte  // 8-byte challenge from the server
	LmChallengeResponse [24]byte // 24-byte LMv2 challenge response
	NtChallengeResponse []byte   // NTProofStr(16) || blob(variable)
}

// NewNTLMv2Response creates a new NTLMv2 response.
//
// Parameters:
//   - username: The username
//   - domain: The domain name
//   - serverChallenge: The 8-byte server challenge
//   - lmChallengeResponse: The 24-byte LMv2 challenge response
//   - ntChallengeResponse: The NT challenge response, NTProofStr(16) || blob
//
// Returns:
//   - A pointer to the new NTLMv2Response
func NewNTLMv2Response(username, domain string, serverChallenge [8]byte, lmChallengeResponse [24]byte, ntChallengeResponse []byte) *NTLMv2Response {
	// Copy the response so a caller reusing its buffer cannot mutate what was
	// captured here.
	nt := make([]byte, len(ntChallengeResponse))
	copy(nt, ntChallengeResponse)

	return &NTLMv2Response{
		Username:            username,
		Domain:              domain,
		ServerChallenge:     serverChallenge,
		LmChallengeResponse: lmChallengeResponse,
		NtChallengeResponse: nt,
	}
}

// NTProofStr returns the 16-byte NTProofStr prefix of the NT challenge response,
// or nil when the response is too short to carry one.
//
// Returns:
//   - The NTProofStr, or nil
func (r *NTLMv2Response) NTProofStr() []byte {
	if len(r.NtChallengeResponse) < NTProofStrLength {
		return nil
	}
	return r.NtChallengeResponse[:NTProofStrLength]
}

// Blob returns the variable-length blob that follows the NTProofStr in the NT
// challenge response, or nil when the response is too short to carry one. The
// blob is what the NTProofStr is computed over, together with the server
// challenge, so cracking needs both.
//
// Returns:
//   - The blob, or nil
func (r *NTLMv2Response) Blob() []byte {
	if len(r.NtChallengeResponse) <= NTProofStrLength {
		return nil
	}
	return r.NtChallengeResponse[NTProofStrLength:]
}

// HashcatString renders the response in the format hashcat mode 5600 expects:
//
//	username::domain:serverchallenge:ntproofstr:blob
//
// The hex is uppercased for consistency with the rest of this repository;
// hashcat parses hex case-insensitively.
//
// Returns:
//   - The hashcat string
//   - An error if the NT challenge response cannot form a mode-5600 line
func (r *NTLMv2Response) HashcatString() (string, error) {
	ntProofStr := r.NTProofStr()
	if ntProofStr == nil {
		return "", fmt.Errorf("NT challenge response is %d bytes, too short to carry the %d-byte NTProofStr",
			len(r.NtChallengeResponse), NTProofStrLength)
	}

	blob := r.Blob()
	if len(blob) == 0 {
		// A response that is exactly the NTProofStr has no blob, and a mode-5600
		// line with an empty final field is not crackable. Report that rather
		// than emitting a malformed line.
		return "", fmt.Errorf("NT challenge response carries no blob after the NTProofStr")
	}

	hashcatString := fmt.Sprintf(
		"%s::%s:%s:%s:%s",
		r.Username,
		r.Domain,
		strings.ToUpper(hex.EncodeToString(r.ServerChallenge[:])),
		strings.ToUpper(hex.EncodeToString(ntProofStr)),
		strings.ToUpper(hex.EncodeToString(blob)),
	)

	return hashcatString, nil
}

// String returns the NTLMv2 response in hashcat mode 5600 form, or the empty
// string when it cannot be rendered.
//
// Returns:
//   - The hashcat string, or ""
func (r *NTLMv2Response) String() string {
	hashcatString, err := r.HashcatString()
	if err != nil {
		return ""
	}
	return hashcatString
}
