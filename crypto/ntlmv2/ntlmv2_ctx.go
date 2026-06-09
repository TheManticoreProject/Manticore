package ntlmv2

import (
	"crypto/hmac"
	"crypto/md5"
	"errors"
	"strings"

	"github.com/TheManticoreProject/Manticore/crypto/lm"
	"github.com/TheManticoreProject/Manticore/crypto/nt"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// NTLMv2 represents the components needed for NTLMv2 authentication
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/5e550938-91d4-459f-b67d-75d70009e3f3
type NTLMv2Ctx struct {
	Domain   string
	Username string
	Password string

	ServerChallenge [8]byte
	ClientChallenge [8]byte

	NTHash [16]byte
	LMHash [16]byte

	ResponseKeyNT [16]byte
	ResponseKeyLM [16]byte
}

// NewNTLMv2CtxWithPassword creates a new NTLMv2 instance with the provided credentials and challenges
//
// Parameters:
//   - domain: The domain name
//   - username: The username
//   - password: The plaintext password
//   - serverChallenge: The 8-byte server challenge
//   - clientChallenge: The 8-byte client challenge
func NewNTLMv2CtxWithPassword(domain, username, password string, serverChallenge, clientChallenge [8]byte) (*NTLMv2Ctx, error) {
	ntHash := nt.NTHash(password)
	return NewNTLMv2CtxWithNTHash(domain, username, ntHash, serverChallenge, clientChallenge)
}

// NewNTLMv2CtxWithNTHash creates a new NTLMv2 instance with the provided credentials and challenges
//
// Parameters:
//   - domain: The domain name
//   - username: The username
//   - nthash: The 16-byte NT hash
//   - serverChallenge: The 8-byte server challenge
//   - clientChallenge: The 8-byte client challenge
func NewNTLMv2CtxWithNTHash(domain, username string, nthash [16]byte, serverChallenge, clientChallenge [8]byte) (*NTLMv2Ctx, error) {
	if len(serverChallenge) != 8 {
		return nil, errors.New("server challenge must be 8 bytes")
	}

	if len(clientChallenge) != 8 {
		return nil, errors.New("client challenge must be 8 bytes")
	}

	ntlm := &NTLMv2Ctx{
		Domain:   domain,
		Username: username,
		Password: "",

		ServerChallenge: serverChallenge,
		ClientChallenge: clientChallenge,

		NTHash: nthash,
		LMHash: lm.LMHash(""),
	}

	// Calculate the ResponseKeyNT = NTOWFv2 = HMAC-MD5(NTHash, UNICODE(Uppercase(User) || Domain)).
	// Per MS-NLMP 3.3.2, only the username is uppercased; the domain is used as-is.
	usernameUpper := strings.ToUpper(username)
	identity := utf16.EncodeUTF16LE(usernameUpper + domain)

	h := hmac.New(md5.New, ntlm.NTHash[:])
	h.Write(identity)
	copy(ntlm.ResponseKeyNT[:], h.Sum(nil))

	return ntlm, nil
}

// ComputeNTChallengeResponse builds the full NTChallengeResponse (NTProofStr || blob)
// as specified in MS-NLMP section 3.3.2.
//
// The blob structure (called "temp" in the spec):
//
//	RespType(1) | HiRespType(1) | Z(2) | Z(4) | Timestamp(8) | ClientChallenge(8) | Z(4) | TargetInfo(var) | Z(4)
//
// targetInfo should already have MsvAvFlags set (use targetinfo.BuildBlobTargetInfo to prepare it).
//
// Parameters:
//   - timestamp: 8-byte Windows FILETIME from MsvAvTimestamp or derived from current time
//   - targetInfo: raw TargetInfo bytes prepared for the blob
//
// Returns:
//   - ntChallengeResponse: NTProofStr(16) || blob(variable)
//   - ntProofStr: the 16-byte NTProofStr (needed for ComputeSessionBaseKey)
//   - error
func (ntlm *NTLMv2Ctx) ComputeNTChallengeResponse(timestamp []byte, targetInfo []byte) ([]byte, []byte, error) {
	if len(timestamp) != 8 {
		return nil, nil, errors.New("timestamp must be 8 bytes")
	}

	// Build the blob (temp)
	blob := make([]byte, 0, 28+len(targetInfo))
	blob = append(blob, 0x01, 0x01)               // RespType, HiRespType
	blob = append(blob, 0x00, 0x00)               // Reserved1
	blob = append(blob, 0x00, 0x00, 0x00, 0x00)   // Reserved2
	blob = append(blob, timestamp...)              // Timestamp (8 bytes)
	blob = append(blob, ntlm.ClientChallenge[:]...) // ClientChallenge (8 bytes)
	blob = append(blob, 0x00, 0x00, 0x00, 0x00)   // Reserved3
	blob = append(blob, targetInfo...)             // TargetInfo (variable)
	blob = append(blob, 0x00, 0x00, 0x00, 0x00)   // Reserved4

	// NTProofStr = HMAC-MD5(ResponseKeyNT, ServerChallenge || blob)
	mac := hmac.New(md5.New, ntlm.ResponseKeyNT[:])
	mac.Write(ntlm.ServerChallenge[:])
	mac.Write(blob)
	ntProofStr := mac.Sum(nil)

	return append(ntProofStr, blob...), ntProofStr, nil
}

// ComputeLMChallengeResponse computes the LmChallengeResponse per MS-NLMP 3.1.5.1.2.
//
// When hasTimestamp is true (MsvAvTimestamp was present in the server TargetInfo),
// the spec requires LmChallengeResponse to be Z(24). Otherwise it is:
//
//	HMAC-MD5(ResponseKeyLM, ServerChallenge || ClientChallenge) || ClientChallenge  (24 bytes)
//
// Parameters:
//   - hasTimestamp: true when MsvAvTimestamp was present in the server's TargetInfo
func (ntlm *NTLMv2Ctx) ComputeLMChallengeResponse(hasTimestamp bool) []byte {
	if hasTimestamp {
		return make([]byte, 24)
	}
	mac := hmac.New(md5.New, ntlm.ResponseKeyNT[:])
	mac.Write(ntlm.ServerChallenge[:])
	mac.Write(ntlm.ClientChallenge[:])
	return append(mac.Sum(nil), ntlm.ClientChallenge[:]...) // 16 + 8 = 24 bytes
}

// ComputeSessionBaseKey derives the SessionBaseKey from the NTProofStr.
//
//	SessionBaseKey = HMAC-MD5(ResponseKeyNT, NTProofStr)
//
// This key is RC4-encrypted with a random session key when KEY_EXCH is negotiated.
//
// Parameters:
//   - ntProofStr: the 16-byte NTProofStr returned by ComputeNTChallengeResponse
func (ntlm *NTLMv2Ctx) ComputeSessionBaseKey(ntProofStr []byte) []byte {
	mac := hmac.New(md5.New, ntlm.ResponseKeyNT[:])
	mac.Write(ntProofStr)
	return mac.Sum(nil)
}

// ComputeResponse computes the combined NTLMv2 challenge response (low-level, explicit params).
// Prefer ComputeNTChallengeResponse and ComputeLMChallengeResponse for new code.
//
// Parameters:
//   - ResponseKeyNT: the 16-byte ResponseKeyNT (NTOWFv2 result)
//   - ResponseKeyLM: the 16-byte ResponseKeyLM (same as ResponseKeyNT for NTLMv2)
//   - ServerChallenge: the 8-byte server challenge
//   - ClientChallenge: the 8-byte client challenge
//   - timestamp: the 8-byte Windows FILETIME timestamp
//   - ServerName: the raw TargetInfo bytes for the blob
//
// Returns:
//   - NTChallengeResponse || LmChallengeResponse concatenated
func (ntlm *NTLMv2Ctx) ComputeResponse(ResponseKeyNT, ResponseKeyLM, ServerChallenge, ClientChallenge, timestamp []byte, ServerName []byte) ([]byte, error) {
	if len(ResponseKeyNT) == 0 && len(ResponseKeyLM) == 0 {
		return []byte{0}, nil
	}

	// Build temp blob:
	// RespType(1) | HiRespType(1) | Z(6) | Timestamp(8) | ClientChallenge(8) | Z(4) | ServerName(var) | Z(4)
	temp := make([]byte, 0)
	temp = append(temp, 0x01)               // Response version
	temp = append(temp, 0x01)               // Hi Response version
	temp = append(temp, make([]byte, 6)...) // Z(6)
	temp = append(temp, timestamp...)       // Timestamp (8 bytes)
	temp = append(temp, ClientChallenge...) // Client challenge
	temp = append(temp, make([]byte, 4)...) // Z(4)
	temp = append(temp, ServerName...)      // Server name (TargetInfo)
	temp = append(temp, make([]byte, 4)...) // Z(4)

	// NTProofStr = HMAC-MD5(ResponseKeyNT, ServerChallenge || temp)
	ntProofStrCtx := hmac.New(md5.New, ResponseKeyNT)
	ntProofStrCtx.Write(ServerChallenge)
	ntProofStrCtx.Write(temp)
	ntProofStr := ntProofStrCtx.Sum(nil)
	NtChallengeResponse := append(ntProofStr, temp...)

	// LmChallengeResponse = HMAC-MD5(ResponseKeyLM, ServerChallenge || ClientChallenge) || ClientChallenge
	lmHmacCtx := hmac.New(md5.New, ResponseKeyLM)
	lmHmacCtx.Write(ServerChallenge)
	lmHmacCtx.Write(ClientChallenge)
	lmChallengeResponse := append(lmHmacCtx.Sum(nil), ClientChallenge...)

	return append(NtChallengeResponse, lmChallengeResponse...), nil
}

// NTOWFv2 computes the NTOWFv2 hash (ResponseKeyNT) for a given password, username, and domain.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/5e550938-91d4-459f-b67d-75d70009e3f3
func NTOWFv2(Passwd, User, UserDomain string) []byte {
	ntHash := nt.NTHash(Passwd)
	userDomainBytes := utf16.EncodeUTF16LE(strings.ToUpper(User) + UserDomain)
	hmacMd5 := hmac.New(md5.New, ntHash[:])
	hmacMd5.Write(userDomainBytes)
	return hmacMd5.Sum(nil)
}

// LMOWFv2 computes the LMOWFv2 hash for a given password, username, and domain
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/5e550938-91d4-459f-b67d-75d70009e3f3
//
// Returns:
//   - The LMOWFv2 hash as a byte slice
//   - An error if the computation fails
func LMOWFv2(Passwd, User, UserDomain string) []byte {
	return NTOWFv2(Passwd, User, UserDomain)
}
