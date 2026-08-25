package ntlmv2

import (
	"crypto/hmac"
	"crypto/md5"
)

// VerifyNTChallengeResponse reports whether an NTLMv2 NT challenge response was
// produced by a party holding responseKeyNT.
//
// It is the acceptor-side counterpart of ComputeNTChallengeResponse. The two
// cannot share an implementation: the computing side builds the blob itself from
// a timestamp and a TargetInfo, whereas a verifier is handed a blob the client
// chose and must recompute the proof over exactly those bytes. Reconstructing the
// blob instead would reject every legitimate client whose blob differs in any
// detail from the one this implementation would have built.
//
//	NTProofStr = HMAC_MD5(ResponseKeyNT, ServerChallenge || blob)
//
// where the response is NTProofStr(16) || blob(variable) ([MS-NLMP] 3.3.2). The
// comparison is constant-time.
//
// Parameters:
//   - responseKeyNT: the NTOWFv2 for the claimed identity, as computed by
//     NewNTLMv2CtxWithNTHash
//   - serverChallenge: the 8-byte challenge the response answers
//   - ntChallengeResponse: the response as received, NTProofStr(16) || blob
//
// Returns:
//   - true when the response verifies against responseKeyNT
func VerifyNTChallengeResponse(responseKeyNT []byte, serverChallenge [8]byte, ntChallengeResponse []byte) bool {
	// A response with no blob cannot be verified: there is nothing for the proof
	// to have been taken over, and an empty-blob proof is not something any
	// client produces.
	if len(responseKeyNT) == 0 || len(ntChallengeResponse) <= NTProofStrLength {
		return false
	}

	ntProofStr := ntChallengeResponse[:NTProofStrLength]
	blob := ntChallengeResponse[NTProofStrLength:]

	mac := hmac.New(md5.New, responseKeyNT)
	mac.Write(serverChallenge[:])
	mac.Write(blob)

	return hmac.Equal(ntProofStr, mac.Sum(nil))
}

// SessionBaseKeyFromResponse derives the SessionBaseKey for a verified NTLMv2
// response: HMAC_MD5(ResponseKeyNT, NTProofStr) ([MS-NLMP] 3.3.2). It is the
// acceptor's route to the key material, taking the NTProofStr from the response
// as received rather than from a computation of its own.
//
// Parameters:
//   - responseKeyNT: the NTOWFv2 for the claimed identity
//   - ntChallengeResponse: the response as received, NTProofStr(16) || blob
//
// Returns:
//   - The 16-byte SessionBaseKey, or nil if the response carries no NTProofStr
func SessionBaseKeyFromResponse(responseKeyNT []byte, ntChallengeResponse []byte) []byte {
	if len(responseKeyNT) == 0 || len(ntChallengeResponse) < NTProofStrLength {
		return nil
	}

	mac := hmac.New(md5.New, responseKeyNT)
	mac.Write(ntChallengeResponse[:NTProofStrLength])
	return mac.Sum(nil)
}
