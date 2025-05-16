package authenticate

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/crypto/ntlmv1"
	"github.com/TheManticoreProject/Manticore/crypto/ntlmv2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/challenge"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/negotiate"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/utils/encoding/utf16"
)

// AuthenticateMessage is the third message in NTLM authentication
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/033d32cc-88f9-4483-9bf2-b273055038ce
type AuthenticateMessage struct {
	// Signature (8 bytes): An 8-byte character array that MUST contain the ASCII string ('N', 'T', 'L', 'M', 'S', 'S', 'P', '\0').
	Signature [8]byte

	// MessageType (4 bytes): A 32-bit unsigned integer that indicates the message type. This field MUST be set to 0x00000003.
	MessageType message.MessageType

	// LmChallengeResponseFields (8 bytes): A field containing LmChallengeResponse information. The field diagram for LmChallengeResponseFields is as follows.
	LmChallengeResponse []byte

	// NtChallengeResponseFields (8 bytes): A field containing NtChallengeResponse information. The field diagram for NtChallengeResponseFields is as follows.
	NtChallengeResponse []byte

	// DomainNameFields (8 bytes): A field containing DomainName information. The field diagram for DomainNameFields is as follows.
	DomainName []byte

	// UserNameFields (8 bytes): A field containing UserName information. The field diagram for the UserNameFields is as follows.
	UserName []byte

	// WorkstationFields (8 bytes): A field containing Workstation information. The field diagram for the WorkstationFields is as follows.
	Workstation []byte

	// EncryptedRandomSessionKeyFields (8 bytes): A field containing EncryptedRandomSessionKey information. The field diagram for EncryptedRandomSessionKeyFields is as follows.
	EncryptedRandomSessionKey []byte

	// NegotiateFlags (4 bytes): In connectionless mode, a NEGOTIATE structure that contains a set of flags (section 2.2.2.5) and represents the conclusion of negotiation—the choices the client has made from the options the server offered in the CHALLENGE_MESSAGE. In connection-oriented mode, a NEGOTIATE structure (section 2.2.2.5) that contains the set of bit flags negotiated in the previous messages.
	NegotiateFlags negotiate.NegotiateFlags

	// Version (8 bytes): A VERSION structure (section 2.2.2.10) that SHOULD be populated only when the NTLMSSP_NEGOTIATE_VERSION flag is set in the NegotiateFlags field; otherwise, it MUST be set to all zero. This structure is used for debugging purposes only. In normal protocol messages, it is ignored and does not affect the NTLM message processing.
	Version version.Version

	// MIC (16 bytes): The message integrity for the NTLM NEGOTIATE_MESSAGE, CHALLENGE_MESSAGE, and AUTHENTICATE_MESSAGE.
	MIC [16]byte
}

// CreateAuthenticateMessage creates an NTLM AUTHENTICATE message
func CreateAuthenticateMessage(challenge *challenge.ChallengeMessage, username, password, domain, workstation string) (*AuthenticateMessage, error) {
	// Determine if we should use Unicode
	useUnicode := (challenge.NegotiateFlags & negotiate.NTLMSSP_NEGOTIATE_UNICODE) != 0

	// Prepare domain, username, and workstation
	var domainBytes, usernameBytes, workstationBytes []byte

	if useUnicode {
		domainBytes = utf16.EncodeUTF16LE(strings.ToUpper(domain))
		usernameBytes = utf16.EncodeUTF16LE(username)
		workstationBytes = utf16.EncodeUTF16LE(strings.ToUpper(workstation))
	} else {
		domainBytes = []byte(strings.ToUpper(domain))
		usernameBytes = []byte(username)
		workstationBytes = []byte(strings.ToUpper(workstation))
	}

	// Calculate NT response
	var lmResponse, ntResponse []byte
	var err error

	if (challenge.NegotiateFlags & negotiate.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY) != 0 {
		// Use NTLMv2
		clientChallenge := [8]byte{}

		_, err = rand.Read(clientChallenge[:])
		if err != nil {
			return nil, err
		}

		ntlmv2, err := ntlmv2.NewNTLMv2WithPassword(domain, username, password, challenge.ServerChallenge, clientChallenge)
		if err != nil {
			return nil, err
		}

		lmResponse, err = ntlmv2.LMResponse()
		if err != nil {
			return nil, err
		}

		ntResponse, err = ntlmv2.NTResponse()
		if err != nil {
			return nil, err
		}
	} else {
		// Use NTLMv1
		ntlmv1, err := ntlmv1.NewNTLMv1WithPassword(domain, username, password, challenge.ServerChallenge[:])
		if err != nil {
			return nil, err
		}

		lmResponse, err = ntlmv1.LMResponse()
		if err != nil {
			return nil, err
		}

		ntResponse, err = ntlmv1.NTResponse()
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	// Prepare session key (empty for now)
	sessionKey := []byte{}

	// Create the AuthenticateMessage struct
	authMsg := &AuthenticateMessage{
		Signature:                 ntlm.NTLM_SIGNATURE,
		MessageType:               message.NTLM_AUTHENTICATE,
		LmChallengeResponse:       lmResponse,
		NtChallengeResponse:       ntResponse,
		DomainName:                domainBytes,
		UserName:                  usernameBytes,
		Workstation:               workstationBytes,
		EncryptedRandomSessionKey: sessionKey,
		NegotiateFlags:            challenge.NegotiateFlags,
	}

	// Set version if needed
	if (challenge.NegotiateFlags & negotiate.NTLMSSP_NEGOTIATE_VERSION) != 0 {
		authMsg.Version = version.DefaultVersion()
	}

	return authMsg, nil
}

// Marshal serializes the AuthenticateMessage into a byte slice
func (msg *AuthenticateMessage) Marshal() ([]byte, error) {
	// Calculate offsets
	headerSize := 88 // Fixed header size including MIC
	lmResponseOffset := headerSize
	ntResponseOffset := lmResponseOffset + len(msg.LmChallengeResponse)
	domainOffset := ntResponseOffset + len(msg.NtChallengeResponse)
	usernameOffset := domainOffset + len(msg.DomainName)
	workstationOffset := usernameOffset + len(msg.UserName)
	sessionKeyOffset := workstationOffset + len(msg.Workstation)

	// Create data
	data := []byte{}

	// Write signature
	data = append(data, msg.Signature[:]...)

	// Write message type
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(msg.MessageType))
	data = append(data, buf...)

	// Write LM response fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.LmChallengeResponse)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.LmChallengeResponse)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(lmResponseOffset))
	data = append(data, buf...)

	// Write NT response fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.NtChallengeResponse)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.NtChallengeResponse)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(ntResponseOffset))
	data = append(data, buf...)

	// Write domain fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.DomainName)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.DomainName)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(domainOffset))
	data = append(data, buf...)

	// Write username fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.UserName)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.UserName)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(usernameOffset))
	data = append(data, buf...)

	// Write workstation fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.Workstation)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.Workstation)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(workstationOffset))
	data = append(data, buf...)

	// Write session key fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.EncryptedRandomSessionKey)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.EncryptedRandomSessionKey)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(sessionKeyOffset))
	data = append(data, buf...)

	// Write negotiate flags
	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(msg.NegotiateFlags))
	data = append(data, buf...)

	// Write version if needed
	if (msg.NegotiateFlags & negotiate.NTLMSSP_NEGOTIATE_VERSION) != 0 {
		byteStream, err := msg.Version.Marshal()
		if err != nil {
			return nil, err
		}
		data = append(data, byteStream...)
	} else {
		// Write 8 bytes of zeros
		data = append(data, make([]byte, 8)...)
	}

	// Write MIC (all zeros for now)
	data = append(data, make([]byte, 16)...)

	// Write payload data
	data = append(data, msg.LmChallengeResponse...)
	data = append(data, msg.NtChallengeResponse...)
	data = append(data, msg.DomainName...)
	data = append(data, msg.UserName...)
	data = append(data, msg.Workstation...)
	data = append(data, msg.EncryptedRandomSessionKey...)

	return data, nil
}

// Unmarshal deserializes a byte slice into an AuthenticateMessage
func (a *AuthenticateMessage) Unmarshal(data []byte) (int, error) {
	if len(data) < 88 {
		return 0, fmt.Errorf("data too short to be a valid AuthenticateMessage")
	}

	msg := &AuthenticateMessage{}

	// Read signature
	copy(msg.Signature[:], data[:8])

	// Read message type
	msg.MessageType = message.MessageType(binary.LittleEndian.Uint32(data[8:12]))

	// Read LM response fields
	lmResponseLen := binary.LittleEndian.Uint16(data[12:14])
	lmResponseOffset := binary.LittleEndian.Uint32(data[16:20])
	msg.LmChallengeResponse = data[lmResponseOffset : lmResponseOffset+uint32(lmResponseLen)]

	// Read NT response fields
	ntResponseLen := binary.LittleEndian.Uint16(data[20:22])
	ntResponseOffset := binary.LittleEndian.Uint32(data[24:28])
	msg.NtChallengeResponse = data[ntResponseOffset : ntResponseOffset+uint32(ntResponseLen)]

	// Read domain fields
	domainLen := binary.LittleEndian.Uint16(data[28:30])
	domainOffset := binary.LittleEndian.Uint32(data[32:36])
	msg.DomainName = data[domainOffset : domainOffset+uint32(domainLen)]

	// Read username fields
	usernameLen := binary.LittleEndian.Uint16(data[36:38])
	usernameOffset := binary.LittleEndian.Uint32(data[40:44])
	msg.UserName = data[usernameOffset : usernameOffset+uint32(usernameLen)]

	// Read workstation fields
	workstationLen := binary.LittleEndian.Uint16(data[44:46])
	workstationOffset := binary.LittleEndian.Uint32(data[48:52])
	msg.Workstation = data[workstationOffset : workstationOffset+uint32(workstationLen)]

	// Read session key fields
	sessionKeyLen := binary.LittleEndian.Uint16(data[52:54])
	sessionKeyOffset := binary.LittleEndian.Uint32(data[56:60])
	msg.EncryptedRandomSessionKey = data[sessionKeyOffset : sessionKeyOffset+uint32(sessionKeyLen)]

	// Read negotiate flags
	msg.NegotiateFlags = negotiate.NegotiateFlags(binary.LittleEndian.Uint32(data[60:64]))

	// Read version if needed
	if (msg.NegotiateFlags & negotiate.NTLMSSP_NEGOTIATE_VERSION) != 0 {
		msg.Version.Unmarshal(data[64:72])
	}

	// Calculate MIC
	micOffset := 72
	micEnd := micOffset + 16
	if micEnd > len(data) {
		return 0, fmt.Errorf("data too short to contain MIC")
	}
	copy(msg.MIC[:], data[micOffset:micEnd])

	return 88, nil
}
