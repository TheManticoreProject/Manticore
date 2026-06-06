package authenticate

import (
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

// CreateAuthenticateMessage creates an NTLM AUTHENTICATE message
func CreateAuthenticateMessage(challenge *challenge.ChallengeMessage, username, password, domain, workstation string) (*AuthenticateMessage, error) {
	// Create the AuthenticateMessage struct
	msg := AuthenticateMessage{}

	msg.NegotiateFlags = challenge.NegotiateFlags

	// Determine if we should use Unicode
	useUnicode := (challenge.NegotiateFlags & flags.NTLMSSP_NEGOTIATE_UNICODE) != 0

	// Prepare domain, username, and workstation
	if useUnicode {
		msg.DomainName = utf16.EncodeUTF16LE(strings.ToUpper(domain))
		msg.UserName = utf16.EncodeUTF16LE(username)
		msg.Workstation = utf16.EncodeUTF16LE(strings.ToUpper(workstation))
	} else {
		msg.DomainName = []byte(strings.ToUpper(domain))
		msg.UserName = []byte(username)
		msg.Workstation = []byte(strings.ToUpper(workstation))
	}

	// Calculate NT response
	var err error

	if (challenge.NegotiateFlags & flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY) != 0 {
		// Use NTLMv2 (MS-NLMP 3.3.2)
		clientChallenge := [8]byte{}
		_, err = rand.Read(clientChallenge[:])
		if err != nil {
			return nil, err
		}

		ctx, err := ntlmv2.NewNTLMv2CtxWithPassword(domain, username, password, challenge.ServerChallenge, clientChallenge)
		if err != nil {
			return nil, err
		}

		// Prepare TargetInfo for the blob: add MsvAvFlags with MIC bit when timestamp is present
		hasTimestamp := targetinfo.HasTimestamp(challenge.TargetInfo)
		blobTargetInfo := targetinfo.BuildBlobTargetInfo(challenge.TargetInfo, hasTimestamp)

		// Use server's MsvAvTimestamp when present; otherwise derive current Windows FILETIME
		timestamp := targetinfo.GetTimestamp(challenge.TargetInfo)
		if len(timestamp) != 8 {
			windowsFiletime := (uint64(time.Now().Unix()) + 116444736000) * 10000000
			timestamp = make([]byte, 8)
			binary.LittleEndian.PutUint64(timestamp, windowsFiletime)
		}

		var ntProofStr []byte
		msg.NtChallengeResponse, ntProofStr, err = ctx.ComputeNTChallengeResponse(timestamp, blobTargetInfo)
		if err != nil {
			return nil, err
		}
		msg.LmChallengeResponse = ctx.ComputeLMChallengeResponse(hasTimestamp)

		// Derive the session key (SessionBaseKey). No key exchange is negotiated here,
		// so the exported session key equals the SessionBaseKey; SMB signing uses it as
		// the MAC key.
		msg.SessionKey = ctx.ComputeSessionBaseKey(ntProofStr)
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

	// Prepare session key (empty for now)
	msg.EncryptedRandomSessionKey = []byte{}

	// Set version if needed
	if (challenge.NegotiateFlags & flags.NTLMSSP_NEGOTIATE_VERSION) != 0 {
		v := version.DefaultVersion()
		msg.Version = &v
	}

	return &msg, nil
}

// Marshal serializes the AuthenticateMessage into a byte slice
func (msg *AuthenticateMessage) Marshal() ([]byte, error) {
	// A 32-bit unsigned integer that defines the offset, in bytes, from
	// the beginning of the AUTHENTICATE_MESSAGE to the entry in Payload
	// Starting at x for length of data section
	offset := 88

	// Write payload data first to compute offsets
	payload := []byte{}

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

	// Write version if needed
	if msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_VERSION) {
		byteStream, err := msg.Version.Marshal()
		if err != nil {
			return nil, err
		}
		marshalledData = append(marshalledData, byteStream...)
	} else {
		// Write 8 bytes of zeros
		marshalledData = append(marshalledData, make([]byte, 8)...)
	}

	// Write MIC (all zeros for now)
	marshalledData = append(marshalledData, make([]byte, 16)...)

	// Write payload
	marshalledData = append(marshalledData, payload...)

	return marshalledData, nil
}

// Unmarshal deserializes a byte slice into an AuthenticateMessage
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
