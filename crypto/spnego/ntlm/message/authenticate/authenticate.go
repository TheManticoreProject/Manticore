package authenticate

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/ntlmv1"
	"github.com/TheManticoreProject/Manticore/crypto/ntlmv2"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/datafields"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/challenge"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/header"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/types"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// AuthenticateMessage is the third message in NTLM authentication
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/033d32cc-88f9-4483-9bf2-b273055038ce
type AuthenticateMessage struct {
	header.Header

	// LmChallengeResponseFields (8 bytes): A field containing LmChallengeResponse information.
	LmChallengeResponseFields datafields.DataFields

	// NtChallengeResponseFields (8 bytes): A field containing NtChallengeResponse information.
	NtChallengeResponseFields datafields.DataFields

	// DomainNameFields (8 bytes): A field containing DomainName information.
	DomainNameFields datafields.DataFields

	// UserNameFields (8 bytes): A field containing UserName information.
	UserNameFields datafields.DataFields

	// WorkstationFields (8 bytes): A field containing Workstation information.
	WorkstationFields datafields.DataFields

	// EncryptedRandomSessionKeyFields (8 bytes): A field containing EncryptedRandomSessionKey information.
	EncryptedRandomSessionKeyFields datafields.DataFields

	// NegotiateFlags (4 bytes): In connectionless mode, a NEGOTIATE structure that contains a set of flags (section 2.2.2.5) and represents the conclusion of negotiation—the choices the client has made from the options the server offered in the CHALLENGE_MESSAGE. In connection-oriented mode, a NEGOTIATE structure (section 2.2.2.5) that contains the set of bit flags negotiated in the previous messages.
	NegotiateFlags flags.NegotiateFlags

	// Version (8 bytes): A VERSION structure (section 2.2.2.10) that SHOULD be populated only when the NTLMSSP_NEGOTIATE_VERSION flag is set in the NegotiateFlags field; otherwise, it MUST be set to all zero. This structure is used for debugging purposes only. In normal protocol messages, it is ignored and does not affect the NTLM message processing.
	Version *version.Version

	// MIC (16 bytes): The message integrity for the NTLM NEGOTIATE_MESSAGE, CHALLENGE_MESSAGE, and AUTHENTICATE_MESSAGE.
	MIC [16]byte

	// NeedsMIC indicates that the server's CHALLENGE carried an MsvAvTimestamp, so
	// the client MUST provide a MIC (MS-NLMP 3.1.5.1.2). When set, the AV_PAIRs in
	// the NtChallengeResponse also carry MsvAvFlags with the MIC-present bit.
	NeedsMIC bool

	// MICOffset is the byte offset the MIC was read from, set by Unmarshal when a
	// MIC was present. A verifier needs it to zero the field in the bytes as
	// received: re-marshalling to recompute the digest would only reproduce the
	// original if this implementation encodes the message identically to whoever
	// sent it, which is not something a verifier can assume.
	MICOffset int

	// Payload section

	// LmChallengeResponse (16 bytes): A payload containing LmChallengeResponse data.
	LmChallengeResponse []byte

	// NtChallengeResponse (16 bytes): A field containing NtChallengeResponse data.
	NtChallengeResponse []byte

	// DomainName (variable): A field containing DomainName data.
	DomainName []byte

	// UserName (variable): A field containing UserName data.
	UserName []byte

	// Workstation (variable): A field containing Workstation data.
	Workstation []byte

	// EncryptedRandomSessionKey (variable): A field containing EncryptedRandomSessionKey data.
	EncryptedRandomSessionKey []byte

	// SessionKey (16 bytes): The exported session key derived during authentication.
	// This is not transmitted on the wire; it is retained so callers (e.g. SMB message
	// signing) can use it as the MAC key. When no key exchange is negotiated it equals
	// the NTLMv2 SessionBaseKey. It is nil for the NTLMv1 path.
	SessionKey []byte
}

// CreateAuthenticateMessage creates an NTLM AUTHENTICATE message from a cleartext
// password. When the server negotiated extended session security it uses NTLMv2;
// otherwise it falls back to NTLMv1.
func CreateAuthenticateMessage(challenge *challenge.ChallengeMessage, username, password, domain, workstation string) (*AuthenticateMessage, error) {
	var v2 *ntlmv2.NTLMv2Ctx
	if (challenge.NegotiateFlags & flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY) != 0 {
		clientChallenge := [8]byte{}
		if _, err := rand.Read(clientChallenge[:]); err != nil {
			return nil, err
		}
		var err error
		v2, err = ntlmv2.NewNTLMv2CtxWithPassword(domain, username, password, challenge.ServerChallenge, clientChallenge)
		if err != nil {
			return nil, err
		}
	}
	return newAuthenticateMessage(challenge, username, password, domain, workstation, v2)
}

// CreateAuthenticateMessageWithNTHash creates an NTLM AUTHENTICATE message from an NT
// hash instead of a cleartext password (pass-the-hash). It requires the server to have
// negotiated extended session security, as the derived session key needed for RPC/SMB
// signing and sealing only exists on the NTLMv2 path; NTLMv1 pass-the-hash is not
// supported.
func CreateAuthenticateMessageWithNTHash(challenge *challenge.ChallengeMessage, username string, ntHash [16]byte, domain, workstation string) (*AuthenticateMessage, error) {
	if (challenge.NegotiateFlags & flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY) == 0 {
		return nil, fmt.Errorf("pass-the-hash requires NTLMv2: server did not negotiate extended session security")
	}
	clientChallenge := [8]byte{}
	if _, err := rand.Read(clientChallenge[:]); err != nil {
		return nil, err
	}
	v2, err := ntlmv2.NewNTLMv2CtxWithNTHash(domain, username, ntHash, challenge.ServerChallenge, clientChallenge)
	if err != nil {
		return nil, err
	}
	return newAuthenticateMessage(challenge, username, "", domain, workstation, v2)
}

// newAuthenticateMessage builds the AUTHENTICATE message shared by the password and
// pass-the-hash entry points. When v2 is non-nil the NTLMv2 (extended session security)
// responses and session key are computed from it; otherwise it falls back to NTLMv1,
// which is driven by the cleartext password.
func newAuthenticateMessage(challenge *challenge.ChallengeMessage, username, password, domain, workstation string, v2 *ntlmv2.NTLMv2Ctx) (*AuthenticateMessage, error) {
	// Create the AuthenticateMessage struct
	msg := AuthenticateMessage{}

	// The AUTHENTICATE carries the client's negotiated flags, not an echo of the
	// server's CHALLENGE flags (which include server-only bits such as the target
	// type and a server VERSION). This matches what the Windows client sends and
	// what Windows Server 2016 (signing required) accepts.
	msg.NegotiateFlags = flags.NTLMSSP_NEGOTIATE_UNICODE |
		flags.NTLMSSP_REQUEST_TARGET |
		flags.NTLMSSP_NEGOTIATE_SIGN |
		flags.NTLMSSP_NEGOTIATE_SEAL |
		flags.NTLMSSP_NEGOTIATE_NTLM |
		flags.NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
		flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
		flags.NTLMSSP_NEGOTIATE_TARGET_INFO |
		flags.NTLMSSP_NEGOTIATE_128 |
		flags.NTLMSSP_NEGOTIATE_56

	// Determine if we should use Unicode
	useUnicode := (challenge.NegotiateFlags & flags.NTLMSSP_NEGOTIATE_UNICODE) != 0

	// Prepare domain, username, and workstation.
	//
	// DomainName MUST be carried with its original case: it has to match, byte for
	// byte, the domain folded into NTOWFv2 below (NTLMv2 uppercases only the username,
	// never the domain — MS-NLMP 3.3.2). Uppercasing it here while NTOWFv2 used the
	// caller's case made the server recompute a different NTProofStr and reject any
	// mixed/lower-case domain (e.g. an FQDN like "host.example.local") with
	// STATUS_LOGON_FAILURE; it only worked when the domain was already upper-case or
	// empty.
	if useUnicode {
		msg.DomainName = utf16.EncodeUTF16LE(domain)
		msg.UserName = utf16.EncodeUTF16LE(username)
		msg.Workstation = utf16.EncodeUTF16LE(strings.ToUpper(workstation))
	} else {
		msg.DomainName = []byte(domain)
		msg.UserName = []byte(username)
		msg.Workstation = []byte(strings.ToUpper(workstation))
	}

	// Calculate NT response
	var err error

	if v2 != nil {
		// Use NTLMv2 (MS-NLMP 3.3.2)

		// Prepare TargetInfo for the blob: add the MsvAvTargetName (SPN) AVPair, as
		// the Windows client does. This server (Server 2016, signing required)
		// requires the SPN rather than an MsvAvFlags MIC, so no MIC is emitted.
		blobTargetInfo := targetinfo.BuildBlobTargetInfo(challenge.TargetInfo)
		msg.NeedsMIC = false

		// Use server's MsvAvTimestamp when present; otherwise derive current Windows FILETIME
		timestamp := targetinfo.GetTimestamp(challenge.TargetInfo)
		if len(timestamp) != 8 {
			windowsFiletime := (uint64(time.Now().Unix()) + 116444736000) * 10000000
			timestamp = make([]byte, 8)
			binary.LittleEndian.PutUint64(timestamp, windowsFiletime)
		}

		var ntProofStr []byte
		msg.NtChallengeResponse, ntProofStr, err = v2.ComputeNTChallengeResponse(timestamp, blobTargetInfo)
		if err != nil {
			return nil, err
		}
		// Send a real LMv2 response (the Windows client does, even with a timestamp).
		msg.LmChallengeResponse = v2.ComputeLMChallengeResponse(false)

		// EXPERIMENT: no key exchange. The exported session key (the SMB signing MAC
		// key) equals the SessionBaseKey.
		msg.SessionKey = v2.ComputeSessionBaseKey(ntProofStr)
	} else {
		// Use NTLMv1
		ntlmv1Ctx, err := ntlmv1.NewNTLMv1CtxWithPassword(domain, username, password, challenge.ServerChallenge)
		if err != nil {
			return nil, err
		}

		response, err := ntlmv1Ctx.ComputeResponse()
		if err != nil {
			return nil, err
		}

		lmChallengeResponse := response.GetLmChallengeResponse()
		msg.LmChallengeResponse = lmChallengeResponse[:]
		if err != nil {
			return nil, err
		}

		ntChallengeResponse := response.GetNtChallengeResponse()
		msg.NtChallengeResponse = ntChallengeResponse[:]
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	// The NTLMv2 path sets EncryptedRandomSessionKey via key exchange above; only
	// default it to empty when it was not produced (e.g. the NTLMv1 path).
	if msg.EncryptedRandomSessionKey == nil {
		msg.EncryptedRandomSessionKey = []byte{}
	}

	// Set version if needed
	if (challenge.NegotiateFlags & flags.NTLMSSP_NEGOTIATE_VERSION) != 0 {
		v := version.DefaultVersion()
		msg.Version = &v
	}

	return &msg, nil
}

// ComputeMIC computes the AUTHENTICATE_MESSAGE message integrity code and stores
// it in the MIC field (MS-NLMP 3.1.5.1.2). The MIC is
// HMAC_MD5(ExportedSessionKey, NEGOTIATE_MESSAGE || CHALLENGE_MESSAGE ||
// AUTHENTICATE_MESSAGE), where the AUTHENTICATE_MESSAGE is serialized with a
// zeroed MIC field. When no key exchange is negotiated, the exported session key
// equals the SessionBaseKey retained in msg.SessionKey.
//
// It must be called after the message is fully populated and only when NeedsMIC
// is set (the server's challenge carried an MsvAvTimestamp).
func (msg *AuthenticateMessage) ComputeMIC(negotiateMessage, challengeMessage []byte) error {
	if len(msg.SessionKey) == 0 {
		return fmt.Errorf("cannot compute NTLM MIC: no session key derived")
	}

	// Serialize with a zeroed MIC field so the hash covers a zero placeholder.
	msg.MIC = [16]byte{}
	authenticateMessage, err := msg.Marshal()
	if err != nil {
		return err
	}

	mac := hmac.New(md5.New, msg.SessionKey)
	mac.Write(negotiateMessage)
	mac.Write(challengeMessage)
	mac.Write(authenticateMessage)
	sum := mac.Sum(nil)
	copy(msg.MIC[:], sum[:16])

	return nil
}

// VerifyMIC reports whether the message integrity code carried by the message
// verifies against an exported session key, the acceptor-side counterpart of
// ComputeMIC.
//
// The digest is taken over the message as received, with only the 16 MIC bytes
// zeroed, rather than over a re-marshalling of the parsed fields: the MIC covers
// the sender's exact encoding, and any difference in how this implementation
// would lay the message out would produce a mismatch that looks identical to a
// forged MIC.
//
//	MIC = HMAC_MD5(ExportedSessionKey, NEGOTIATE || CHALLENGE || AUTHENTICATE)
//
// The comparison is constant-time.
//
// Parameters:
//   - exportedSessionKey: the key derived from the verified NT response
//   - negotiateMessage: the raw NEGOTIATE_MESSAGE of this exchange
//   - challengeMessage: the raw CHALLENGE_MESSAGE of this exchange
//   - rawAuthenticate: this message exactly as received
//
// Returns:
//   - true when the MIC verifies
func (msg *AuthenticateMessage) VerifyMIC(exportedSessionKey, negotiateMessage, challengeMessage, rawAuthenticate []byte) bool {
	if len(exportedSessionKey) == 0 || !msg.NeedsMIC {
		return false
	}
	if msg.MICOffset <= 0 || msg.MICOffset+MICLength > len(rawAuthenticate) {
		return false
	}

	// Zero the MIC field in a copy, so the caller's buffer is untouched.
	zeroed := make([]byte, len(rawAuthenticate))
	copy(zeroed, rawAuthenticate)
	for i := msg.MICOffset; i < msg.MICOffset+MICLength; i++ {
		zeroed[i] = 0x00
	}

	mac := hmac.New(md5.New, exportedSessionKey)
	mac.Write(negotiateMessage)
	mac.Write(challengeMessage)
	mac.Write(zeroed)

	return hmac.Equal(msg.MIC[:], mac.Sum(nil))
}

// Marshal serializes the AuthenticateMessage into a byte slice
func (msg *AuthenticateMessage) Marshal() ([]byte, error) {
	// Offset, in bytes, from the start of the AUTHENTICATE_MESSAGE to the payload.
	// After the 64-byte fixed header, the optional Version (8 bytes) is present only
	// when NTLMSSP_NEGOTIATE_VERSION is set, and the optional MIC (16 bytes) only
	// when a MIC is in use. The Windows client omits both when not in use; emitting
	// spurious zero-filled Version/MIC fields is rejected by Windows Server 2016.
	hasVersion := msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_VERSION)
	offset := 64
	if hasVersion {
		offset += 8
	}
	if msg.NeedsMIC {
		offset += 16
	}

	// Write payload data first to compute offsets. The payload is laid out in the
	// canonical Windows-client order — DomainName, UserName, Workstation,
	// LmChallengeResponse, NtChallengeResponse, EncryptedRandomSessionKey — which
	// Windows Server 2016 expects (it rejects an out-of-order payload with
	// STATUS_INVALID_PARAMETER even though the field offsets are self-describing).
	payload := []byte{}

	// Domain name
	msg.DomainNameFields.Len = uint16(len(msg.DomainName))
	msg.DomainNameFields.MaxLen = uint16(len(msg.DomainName))
	msg.DomainNameFields.BufferOffset = uint32(offset)
	offset += len(msg.DomainName)
	payload = append(payload, msg.DomainName...)

	// User name
	msg.UserNameFields.Len = uint16(len(msg.UserName))
	msg.UserNameFields.MaxLen = uint16(len(msg.UserName))
	msg.UserNameFields.BufferOffset = uint32(offset)
	offset += len(msg.UserName)
	payload = append(payload, msg.UserName...)

	// Workstation
	msg.WorkstationFields.Len = uint16(len(msg.Workstation))
	msg.WorkstationFields.MaxLen = uint16(len(msg.Workstation))
	msg.WorkstationFields.BufferOffset = uint32(offset)
	offset += len(msg.Workstation)
	payload = append(payload, msg.Workstation...)

	// LM response
	msg.LmChallengeResponseFields.Len = uint16(len(msg.LmChallengeResponse))
	msg.LmChallengeResponseFields.MaxLen = uint16(len(msg.LmChallengeResponse))
	msg.LmChallengeResponseFields.BufferOffset = uint32(offset)
	offset += len(msg.LmChallengeResponse)
	payload = append(payload, msg.LmChallengeResponse...)

	// NT response
	msg.NtChallengeResponseFields.Len = uint16(len(msg.NtChallengeResponse))
	msg.NtChallengeResponseFields.MaxLen = uint16(len(msg.NtChallengeResponse))
	msg.NtChallengeResponseFields.BufferOffset = uint32(offset)
	offset += len(msg.NtChallengeResponse)
	payload = append(payload, msg.NtChallengeResponse...)

	// Encrypted random session key
	msg.EncryptedRandomSessionKeyFields.Len = uint16(len(msg.EncryptedRandomSessionKey))
	msg.EncryptedRandomSessionKeyFields.MaxLen = uint16(len(msg.EncryptedRandomSessionKey))
	msg.EncryptedRandomSessionKeyFields.BufferOffset = uint32(offset)
	offset += len(msg.EncryptedRandomSessionKey)
	payload = append(payload, msg.EncryptedRandomSessionKey...)

	// Write data section
	marshalledData := []byte{}

	// Write header
	msg.Header.MessageType = types.MESSAGE_TYPE_AUTHENTICATE
	msg.Header.Signature = header.NTLM_SIGNATURE
	marshalledHeader, err := msg.Header.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, marshalledHeader...)

	// Write LM response fields
	lmChallengeResponseFieldsBytes, err := msg.LmChallengeResponseFields.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, lmChallengeResponseFieldsBytes...)

	// Write NT response fields
	ntChallengeResponseFieldsBytes, err := msg.NtChallengeResponseFields.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, ntChallengeResponseFieldsBytes...)

	// Write domain fields
	domainNameFieldsBytes, err := msg.DomainNameFields.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, domainNameFieldsBytes...)

	// Write username fields
	userNameFieldsBytes, err := msg.UserNameFields.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, userNameFieldsBytes...)

	// Write workstation fields
	workstationFieldsBytes, err := msg.WorkstationFields.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, workstationFieldsBytes...)

	// Write session key fields
	encryptedRandomSessionKeyFieldsBytes, err := msg.EncryptedRandomSessionKeyFields.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, encryptedRandomSessionKeyFieldsBytes...)

	// Write negotiate flags
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(msg.NegotiateFlags))
	marshalledData = append(marshalledData, buf...)

	// Write the Version field only when NTLMSSP_NEGOTIATE_VERSION is set; otherwise
	// it is omitted entirely (not zero-filled).
	if hasVersion {
		byteStream, err := msg.Version.Marshal()
		if err != nil {
			return nil, err
		}
		marshalledData = append(marshalledData, byteStream...)
	}

	// Write the 16-byte MIC field only when a MIC is in use. When NeedsMIC is set,
	// the field is zero while ComputeMIC marshals the message to hash it, then holds
	// the computed MIC.
	if msg.NeedsMIC {
		marshalledData = append(marshalledData, msg.MIC[:]...)
	}

	// Write payload
	marshalledData = append(marshalledData, payload...)

	return marshalledData, nil
}

// Unmarshal deserializes a byte slice into an AuthenticateMessage
// MICLength is the length in bytes of the AUTHENTICATE_MESSAGE message integrity
// code.
const MICLength = 16

// payloadStart returns the lowest buffer offset among the payload fields, which
// is where the fixed part of the message ends. A field with a zero length carries
// no payload and its offset is not meaningful, so it is skipped.
//
// It is only valid once the field descriptors have been read.
func (msg *AuthenticateMessage) payloadStart() int {
	start := 0
	consider := func(length uint16, offset uint32) {
		if length == 0 {
			return
		}
		if start == 0 || int(offset) < start {
			start = int(offset)
		}
	}
	consider(msg.LmChallengeResponseFields.Len, msg.LmChallengeResponseFields.BufferOffset)
	consider(msg.NtChallengeResponseFields.Len, msg.NtChallengeResponseFields.BufferOffset)
	consider(msg.DomainNameFields.Len, msg.DomainNameFields.BufferOffset)
	consider(msg.UserNameFields.Len, msg.UserNameFields.BufferOffset)
	consider(msg.WorkstationFields.Len, msg.WorkstationFields.BufferOffset)
	consider(msg.EncryptedRandomSessionKeyFields.Len, msg.EncryptedRandomSessionKeyFields.BufferOffset)
	return start
}

func (msg *AuthenticateMessage) Unmarshal(data []byte) (int, error) {
	totalBytesRead := 0

	if len(data) < 88 {
		return 0, fmt.Errorf("data too short to be a valid AuthenticateMessage")
	}

	// Read header
	bytesRead, err := msg.Header.Unmarshal(data)
	if err != nil {
		return 0, err
	}
	if err := msg.Header.Expect(types.MESSAGE_TYPE_AUTHENTICATE); err != nil {
		return 0, err
	}
	totalBytesRead += bytesRead

	// Read LM response fields
	if len(data) < 16 {
		return 0, fmt.Errorf("data too short to read LmChallengeResponseFields in AuthenticateMessage")
	}
	bytesRead, err = msg.LmChallengeResponseFields.Unmarshal(data[totalBytesRead:])
	if err != nil {
		return 0, err
	}
	totalBytesRead += bytesRead

	// Read NT response fields
	if len(data) < 16 {
		return 0, fmt.Errorf("data too short to read NtChallengeResponseFields in AuthenticateMessage")
	}
	bytesRead, err = msg.NtChallengeResponseFields.Unmarshal(data[totalBytesRead:])
	if err != nil {
		return 0, err
	}
	totalBytesRead += bytesRead

	// Read domain fields
	if len(data) < 16 {
		return 0, fmt.Errorf("data too short to read DomainNameFields in AuthenticateMessage")
	}
	bytesRead, err = msg.DomainNameFields.Unmarshal(data[totalBytesRead:])
	if err != nil {
		return 0, err
	}
	totalBytesRead += bytesRead

	// Read username fields
	if len(data) < 16 {
		return 0, fmt.Errorf("data too short to read UserNameFields in AuthenticateMessage")
	}
	bytesRead, err = msg.UserNameFields.Unmarshal(data[totalBytesRead:])
	if err != nil {
		return 0, err
	}
	totalBytesRead += bytesRead

	// Read workstation fields
	if len(data) < 16 {
		return 0, fmt.Errorf("data too short to read WorkstationFields in AuthenticateMessage")
	}
	bytesRead, err = msg.WorkstationFields.Unmarshal(data[totalBytesRead:])
	if err != nil {
		return 0, err
	}
	totalBytesRead += bytesRead

	// Read session key fields
	if len(data) < 16 {
		return 0, fmt.Errorf("data too short to read EncryptedRandomSessionKeyFields in AuthenticateMessage")
	}
	bytesRead, err = msg.EncryptedRandomSessionKeyFields.Unmarshal(data[totalBytesRead:])
	if err != nil {
		return 0, err
	}
	totalBytesRead += bytesRead

	// Read negotiate flags
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short to read NegotiateFlags in AuthenticateMessage")
	}
	msg.NegotiateFlags = flags.NegotiateFlags(binary.LittleEndian.Uint32(data[totalBytesRead : totalBytesRead+4]))
	totalBytesRead += 4

	// Read version if needed
	if (msg.NegotiateFlags & flags.NTLMSSP_NEGOTIATE_VERSION) != 0 {
		if (totalBytesRead + 8) > len(data) {
			return 0, fmt.Errorf("data too short to read Version in AuthenticateMessage")
		}
		if msg.Version == nil {
			msg.Version = &version.Version{}
		}
		bytesRead, err = msg.Version.Unmarshal(data[totalBytesRead : totalBytesRead+8])
		if err != nil {
			return 0, err
		}
		totalBytesRead += bytesRead
	} else {
		msg.Version = nil
		// Read 8 bytes of zeros
		totalBytesRead += 8
	}

	// Read the MIC when one is present.
	//
	// The MIC is not self-describing: nothing in the fixed header announces it. It
	// occupies the 16 bytes immediately before the payload, so it is located from
	// where the payload starts rather than from the cursor above. That matters,
	// because two layouts appear in the wild and they put the payload in different
	// places: [MS-NLMP] treats Version as always present (payload at 72 without a
	// MIC, 88 with one), while this implementation's Marshal omits Version when
	// NTLMSSP_NEGOTIATE_VERSION is clear (payload at 64, or 80 with a MIC).
	// Deriving the offset from the payload handles both, and a MIC is present in
	// either layout exactly when the payload starts at 80 or beyond.
	//
	// Without this the MIC a client sent was silently discarded, so an acceptor had
	// nothing to verify and a tampered MIC was indistinguishable from a correct
	// one.
	if payloadStart := msg.payloadStart(); payloadStart >= 80 {
		micOffset := payloadStart - MICLength
		if micOffset < 64 || payloadStart > len(data) {
			return 0, fmt.Errorf("data too short to read MIC in AuthenticateMessage")
		}
		copy(msg.MIC[:], data[micOffset:payloadStart])
		msg.NeedsMIC = true
		msg.MICOffset = micOffset
	}

	// Read payload section

	// LM response
	if int(msg.LmChallengeResponseFields.BufferOffset)+int(msg.LmChallengeResponseFields.Len) > len(data) {
		return 0, fmt.Errorf("data too short to read LmChallengeResponse in payload section in AuthenticateMessage")
	}
	msg.LmChallengeResponse = data[int(msg.LmChallengeResponseFields.BufferOffset) : int(msg.LmChallengeResponseFields.BufferOffset)+int(msg.LmChallengeResponseFields.Len)]
	totalBytesRead += int(msg.LmChallengeResponseFields.Len)

	// NT response
	if int(msg.NtChallengeResponseFields.BufferOffset)+int(msg.NtChallengeResponseFields.Len) > len(data) {
		return 0, fmt.Errorf("data too short to read NtChallengeResponse in payload section in AuthenticateMessage")
	}
	msg.NtChallengeResponse = data[int(msg.NtChallengeResponseFields.BufferOffset) : int(msg.NtChallengeResponseFields.BufferOffset)+int(msg.NtChallengeResponseFields.Len)]
	totalBytesRead += int(msg.NtChallengeResponseFields.Len)

	// Domain name
	if int(msg.DomainNameFields.BufferOffset)+int(msg.DomainNameFields.Len) > len(data) {
		return 0, fmt.Errorf("data too short to read DomainName in payload section in AuthenticateMessage")
	}
	msg.DomainName = data[int(msg.DomainNameFields.BufferOffset) : int(msg.DomainNameFields.BufferOffset)+int(msg.DomainNameFields.Len)]
	totalBytesRead += int(msg.DomainNameFields.Len)

	// User name
	if int(msg.UserNameFields.BufferOffset)+int(msg.UserNameFields.Len) > len(data) {
		return 0, fmt.Errorf("data too short to read UserName in payload section in AuthenticateMessage")
	}
	msg.UserName = data[int(msg.UserNameFields.BufferOffset) : int(msg.UserNameFields.BufferOffset)+int(msg.UserNameFields.Len)]
	totalBytesRead += int(msg.UserNameFields.Len)

	// Workstation
	if int(msg.WorkstationFields.BufferOffset)+int(msg.WorkstationFields.Len) > len(data) {
		return 0, fmt.Errorf("data too short to read Workstation in payload section in AuthenticateMessage")
	}
	msg.Workstation = data[int(msg.WorkstationFields.BufferOffset) : int(msg.WorkstationFields.BufferOffset)+int(msg.WorkstationFields.Len)]
	totalBytesRead += int(msg.WorkstationFields.Len)

	// Encrypted random session key
	if int(msg.EncryptedRandomSessionKeyFields.BufferOffset)+int(msg.EncryptedRandomSessionKeyFields.Len) > len(data) {
		return 0, fmt.Errorf("data too short to read EncryptedRandomSessionKey in payload section in AuthenticateMessage")
	}
	msg.EncryptedRandomSessionKey = data[int(msg.EncryptedRandomSessionKeyFields.BufferOffset) : int(msg.EncryptedRandomSessionKeyFields.BufferOffset)+int(msg.EncryptedRandomSessionKeyFields.Len)]
	totalBytesRead += int(msg.EncryptedRandomSessionKeyFields.Len)

	return totalBytesRead, nil
}

// GetMessageType returns the message type of the AuthenticateMessage
func (msg *AuthenticateMessage) GetMessageType() uint32 {
	return uint32(msg.Header.MessageType)
}
