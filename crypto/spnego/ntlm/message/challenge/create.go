package challenge

import (
	"crypto/rand"
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/header"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/types"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// acceptorSupportedFlags are the options this implementation will agree to when a
// client offers them. The CHALLENGE advertises the intersection of the client's
// offer with this set, per [MS-NLMP] 3.2.5.1.1, so a client cannot select
// something the acceptor does not implement, and the acceptor does not assert
// something the client did not ask for.
const acceptorSupportedFlags = flags.NTLMSSP_NEGOTIATE_UNICODE |
	flags.NTLMSSP_NEGOTIATE_OEM |
	flags.NTLMSSP_REQUEST_TARGET |
	flags.NTLMSSP_NEGOTIATE_SIGN |
	flags.NTLMSSP_NEGOTIATE_SEAL |
	flags.NTLMSSP_NEGOTIATE_NTLM |
	flags.NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
	flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
	flags.NTLMSSP_NEGOTIATE_TARGET_INFO |
	flags.NTLMSSP_NEGOTIATE_KEY_EXCH |
	flags.NTLMSSP_NEGOTIATE_128 |
	flags.NTLMSSP_NEGOTIATE_56

// TargetType selects which target-type bit the CHALLENGE asserts, describing what
// the TargetName names.
type TargetType int

const (
	// TargetTypeDomain declares the TargetName to be a domain name.
	TargetTypeDomain TargetType = iota
	// TargetTypeServer declares the TargetName to be a server name.
	TargetTypeServer
	// TargetTypeNone asserts neither bit.
	TargetTypeNone
)

// NewServerChallenge returns a fresh 8-byte challenge from the system's
// cryptographic random source.
//
// The challenge is the only thing standing between a captured response and an
// offline replay, so it must be unpredictable and must not be reused across
// exchanges.
//
// Returns:
//   - The challenge
//   - An error if the random source fails
func NewServerChallenge() ([8]byte, error) {
	var challenge [8]byte
	if _, err := rand.Read(challenge[:]); err != nil {
		return challenge, fmt.Errorf("failed to generate a server challenge: %v", err)
	}
	return challenge, nil
}

// CreateChallengeMessage builds the CHALLENGE_MESSAGE answering a client's
// NEGOTIATE_MESSAGE, the acceptor-side counterpart of CreateNegotiateMessage.
//
// The negotiated flags are the intersection of what the client offered with what
// this implementation supports, with three adjustments the specification requires
// of an acceptor:
//
//   - Unicode wins over OEM when the client offered both, because every string in
//     the exchange is then UTF-16LE and mixing the two is what produces an
//     identity that fails to verify.
//   - REQUEST_TARGET and NEGOTIATE_TARGET_INFO are asserted whenever a TargetName
//     or TargetInfo is supplied. A Windows client abandons the exchange when it
//     receives TargetInfo that the flags do not announce.
//   - NEGOTIATE_VERSION is asserted only when a Version is supplied, since the
//     field is otherwise required to be zero.
//
// Parameters:
//   - neg: the client's NEGOTIATE_MESSAGE
//   - serverChallenge: the 8-byte challenge to issue
//   - targetName: the name to advertise, or "" to omit it
//   - targetType: which target-type bit to assert
//   - targetInfo: the AV_PAIR list to advertise, or nil to omit it
//   - v: the version to advertise, or nil to omit it
//
// Returns:
//   - The CHALLENGE_MESSAGE
//   - An error if the client's offer cannot be answered
func CreateChallengeMessage(
	neg *negotiate.NegotiateMessage,
	serverChallenge [8]byte,
	targetName string,
	targetType TargetType,
	targetInfo []byte,
	v *version.Version,
) (*ChallengeMessage, error) {
	if neg == nil {
		return nil, fmt.Errorf("cannot answer a nil NEGOTIATE message")
	}

	negotiated := neg.NegotiateFlags & acceptorSupportedFlags

	// A client that offers neither character set has offered nothing this
	// implementation can speak, and there is no sane default: picking one would
	// mean encoding the identity differently from how the client folded it into
	// its response.
	if negotiated&(flags.NTLMSSP_NEGOTIATE_UNICODE|flags.NTLMSSP_NEGOTIATE_OEM) == 0 {
		return nil, fmt.Errorf("client offered neither NTLMSSP_NEGOTIATE_UNICODE nor NTLMSSP_NEGOTIATE_OEM")
	}
	// Exactly one character set is negotiated, and Unicode is preferred.
	if negotiated.HasFlag(flags.NTLMSSP_NEGOTIATE_UNICODE) {
		negotiated &= ^flags.NTLMSSP_NEGOTIATE_OEM
	}

	msg := &ChallengeMessage{
		Header: header.Header{
			Signature:   header.NTLM_SIGNATURE,
			MessageType: types.MESSAGE_TYPE_CHALLENGE,
		},
		ServerChallenge: serverChallenge,
	}

	if targetName != "" {
		if negotiated.HasFlag(flags.NTLMSSP_NEGOTIATE_UNICODE) {
			msg.TargetName = utf16.EncodeUTF16LE(targetName)
		} else {
			msg.TargetName = []byte(targetName)
		}
		negotiated |= flags.NTLMSSP_REQUEST_TARGET

		switch targetType {
		case TargetTypeDomain:
			negotiated |= flags.NTLMSSP_TARGET_TYPE_DOMAIN
		case TargetTypeServer:
			negotiated |= flags.NTLMSSP_TARGET_TYPE_SERVER
		}
	}

	if len(targetInfo) > 0 {
		msg.TargetInfo = targetInfo
		negotiated |= flags.NTLMSSP_NEGOTIATE_TARGET_INFO
		// TargetInfo is only meaningful alongside extended session security, and
		// a client that sends a v2 response expects the bit set.
		negotiated |= flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY
	}

	if v != nil {
		msg.Version = v
		negotiated |= flags.NTLMSSP_NEGOTIATE_VERSION
	} else {
		negotiated &= ^flags.NTLMSSP_NEGOTIATE_VERSION
	}

	msg.NegotiateFlags = negotiated

	return msg, nil
}
