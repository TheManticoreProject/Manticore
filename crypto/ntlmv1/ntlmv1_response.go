package ntlmv1

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// HashcatMode is the hashcat hash-mode number for a NetNTLMv1 response, so a
// caller writing captured material to a file can label it correctly. It is a
// different mode from a NetNTLMv2 response, which hashcat expects in a different
// field layout (see crypto/ntlmv2.HashcatMode).
const HashcatMode = 5500

type NTLMv1Response struct {
	Username string // Username for authentication
	Domain   string // Domain name

	ServerChallenge     [8]byte  // 8-byte challenge from server
	LmChallengeResponse [24]byte // 24-byte LM challenge response
	NtChallengeResponse [24]byte // 24-byte NT challenge response
}

func NewNTLMv1Response(username, domain string, serverChallenge [8]byte, lmChallengeResponse [24]byte, ntChallengeResponse [24]byte) *NTLMv1Response {
	return &NTLMv1Response{
		Username:            username,
		Domain:              domain,
		ServerChallenge:     serverChallenge,
		LmChallengeResponse: lmChallengeResponse,
		NtChallengeResponse: ntChallengeResponse,
	}
}

// GetServerChallenge returns the server challenge
func (r *NTLMv1Response) GetServerChallenge() [8]byte {
	return r.ServerChallenge
}

// SetServerChallenge sets the server challenge
func (r *NTLMv1Response) SetServerChallenge(challenge [8]byte) {
	r.ServerChallenge = challenge
}

// GetLmChallengeResponse returns the LM challenge response
func (r *NTLMv1Response) GetLmChallengeResponse() [24]byte {
	return r.LmChallengeResponse
}

// SetLmChallengeResponse sets the LM challenge response
func (r *NTLMv1Response) SetLmChallengeResponse(response [24]byte) {
	r.LmChallengeResponse = response
}

// GetNtChallengeResponse returns the NT challenge response
func (r *NTLMv1Response) GetNtChallengeResponse() [24]byte {
	return r.NtChallengeResponse
}

// SetNtChallengeResponse sets the NT challenge response
func (r *NTLMv1Response) SetNtChallengeResponse(response [24]byte) {
	r.NtChallengeResponse = response
}

// Equal compares two NTLMv1Response objects for equality
func (r *NTLMv1Response) Equal(other *NTLMv1Response) bool {
	if other == nil {
		return false
	}
	return r.ServerChallenge == other.ServerChallenge &&
		r.LmChallengeResponse == other.LmChallengeResponse &&
		r.NtChallengeResponse == other.NtChallengeResponse
}

// HashcatString converts the NTLMv1 response to a Hashcat string
//
// Returns:
//   - The Hashcat string
//   - An error if the conversion fails
func (r *NTLMv1Response) HashcatString() (string, error) {
	hashcatString := fmt.Sprintf(
		"%s::%s:%s:%s:%s",
		r.Username,
		r.Domain,
		strings.ToUpper(hex.EncodeToString(r.LmChallengeResponse[:])),
		strings.ToUpper(hex.EncodeToString(r.NtChallengeResponse[:])),
		strings.ToUpper(hex.EncodeToString(r.ServerChallenge[:])),
	)

	return hashcatString, nil
}

// String returns the NTLMv1 response as a string
//
// Returns:
//   - The NTLMv1 response as a string
//   - An error if the conversion fails
func (r *NTLMv1Response) String() string {
	hashcatString, err := r.HashcatString()
	if err != nil {
		return ""
	}
	return hashcatString
}
