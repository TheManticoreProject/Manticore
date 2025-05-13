package negotiate

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/utils/encoding/utf16"
)

// NegotiateMessage is the first message in NTLM authentication
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/b34032e5-3aae-4bc6-84c3-c6d80eadf7f2
type NegotiateMessage struct {
	// Signature (8 bytes): An 8-byte character array that MUST contain the ASCII string ('N', 'T', 'L', 'M', 'S', 'S', 'P', '\0').
	Signature [8]byte

	// MessageType (4 bytes): A 32-bit unsigned integer that indicates the message type. This field MUST be set to 0x00000001.
	MessageType message.MessageType

	// NegotiateFlags (4 bytes): A NEGOTIATE structure that contains a set of flags, as defined in section 2.2.2.5. The client sets flags to indicate options it supports.
	NegotiateFlags NegotiateFlags

	// DomainNameFields (8 bytes): A field containing DomainName information. The field diagram for DomainNameFields is as follows.
	DomainName []byte

	// WorkstationFields (8 bytes): A field containing WorkstationName information. The field diagram for WorkstationFields is as follows.
	Workstation []byte

	// Version (8 bytes): A VERSION structure (as defined in section 2.2.2.10) that is populated only when the NTLMSSP_NEGOTIATE_VERSION flag is set in the NegotiateFlags field; otherwise, it MUST be set to all zero. This structure SHOULD<6> be used for debugging purposes only. In normal (nondebugging) protocol messages, it is ignored and does not affect the NTLM message processing.
	Version version.Version

	// Payload (variable): A byte-array that contains the data referred to by the DomainNameBufferOffset and WorkstationBufferOffset fields. Payload data can be present in any order within the Payload field, with variable-length padding before or after the data. The data that can be present in the Payload field of this message, in no particular order, are:
	Payload []byte
}

// CreateNegotiateMessage initializes a NegotiateMessage with the given parameters
func CreateNegotiateMessage(domain, workstation string, useUnicode bool) (*NegotiateMessage, error) {
	flags := NTLMSSP_NEGOTIATE_NTLM |
		NTLMSSP_NEGOTIATE_ALWAYS_SIGN |
		NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY |
		NTLMSSP_NEGOTIATE_128 |
		NTLMSSP_NEGOTIATE_56 |
		NTLMSSP_REQUEST_TARGET |
		NTLMSSP_NEGOTIATE_TARGET_INFO |
		NTLMSSP_NEGOTIATE_VERSION

	// Set Unicode flag
	if useUnicode {
		flags |= NTLMSSP_NEGOTIATE_UNICODE
	} else {
		flags |= NTLMSSP_NEGOTIATE_OEM
	}

	var domainBytes, workstationBytes []byte

	if domain != "" {
		flags |= NTLMSSP_NEGOTIATE_OEM_DOMAIN_SUPPLIED
		if useUnicode {
			domainBytes = utf16.EncodeUTF16LE(domain)
		} else {
			domainBytes = []byte(strings.ToUpper(domain))
		}
	}

	if workstation != "" {
		flags |= NTLMSSP_NEGOTIATE_OEM_WORKSTATION_SUPPLIED
		if useUnicode {
			workstationBytes = utf16.EncodeUTF16LE(workstation)
		} else {
			workstationBytes = []byte(strings.ToUpper(workstation))
		}
	}

	return &NegotiateMessage{
		Signature:      ntlm.NTLM_SIGNATURE,
		MessageType:    message.NTLM_NEGOTIATE,
		NegotiateFlags: flags,
		DomainName:     domainBytes,
		Workstation:    workstationBytes,
		Version:        version.DefaultVersion(),
	}, nil
}

// Marshal serializes the NegotiateMessage into a byte slice
func (msg *NegotiateMessage) Marshal() ([]byte, error) {
	// Calculate offsets and sizes
	headerSize := 40 // 8 (signature) + 4 (type) + 4 (flags) + 8 (domain fields) + 8 (workstation fields) + 8 (version)
	domainOffset := headerSize
	workstationOffset := domainOffset + len(msg.DomainName)

	// Create data
	data := []byte{}

	// Write signature
	data = append(data, msg.Signature[:]...)

	// Write message type
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(msg.MessageType))
	data = append(data, buf...)

	// Write negotiate flags
	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(msg.NegotiateFlags))
	data = append(data, buf...)

	// Write domain name fields
	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.DomainName)))
	data = append(data, buf...)

	buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(len(msg.DomainName)))
	data = append(data, buf...)

	buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(domainOffset))
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

	// Write version
	byteStream, err := msg.Version.Marshal()
	if err != nil {
		return nil, err
	}
	data = append(data, byteStream...)

	// Write domain and workstation
	data = append(data, msg.DomainName...)
	data = append(data, msg.Workstation...)

	return data, nil
}

// Unmarshal deserializes a byte slice into a NegotiateMessage
func (msg *NegotiateMessage) Unmarshal(data []byte) error {
	if len(data) < 40 {
		return fmt.Errorf("data too short to be a valid NegotiateMessage")
	}

	copy(msg.Signature[:], data[:8])

	msg.MessageType = message.MessageType(binary.LittleEndian.Uint32(data[8:12]))

	msg.NegotiateFlags = NegotiateFlags(binary.LittleEndian.Uint32(data[12:16]))

	domainLen := binary.LittleEndian.Uint16(data[16:18])
	domainOffset := binary.LittleEndian.Uint32(data[20:24])

	workstationLen := binary.LittleEndian.Uint16(data[24:26])
	workstationOffset := binary.LittleEndian.Uint32(data[28:32])

	if domainOffset+uint32(domainLen) > uint32(len(data)) || workstationOffset+uint32(workstationLen) > uint32(len(data)) {
		return fmt.Errorf("invalid offsets or lengths in NegotiateMessage")
	}

	msg.DomainName = data[domainOffset : domainOffset+uint32(domainLen)]

	msg.Workstation = data[workstationOffset : workstationOffset+uint32(workstationLen)]

	_, err := msg.Version.Unmarshal(data[32:40])

	return err
}
