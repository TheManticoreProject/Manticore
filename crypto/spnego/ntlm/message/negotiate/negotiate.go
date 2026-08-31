package negotiate

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/datafields"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/header"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/types"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// NegotiateMessage is the first message in NTLM authentication
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/b34032e5-3aae-4bc6-84c3-c6d80eadf7f2
type NegotiateMessage struct {
	header.Header

	// NegotiateFlags (4 bytes): A NEGOTIATE structure that contains a set of flags, as defined in section 2.2.2.5. The client sets flags to indicate options it supports.
	NegotiateFlags flags.NegotiateFlags

	// DomainNameFields (8 bytes): A field containing DomainName information.
	DomainNameFields datafields.DataFields

	// WorkstationFields (8 bytes): A field containing WorkstationName information.
	WorkstationFields datafields.DataFields

	// Version (8 bytes): A VERSION structure (as defined in section 2.2.2.10) that is populated only when the NTLMSSP_NEGOTIATE_VERSION flag is set in the NegotiateFlags field; otherwise, it MUST be set to all zero. This structure SHOULD<6> be used for debugging purposes only. In normal (nondebugging) protocol messages, it is ignored and does not affect the NTLM message processing.
	Version *version.Version

	// Payload section

	// DomainName (variable): A field containing DomainName data.
	DomainName []byte

	// Workstation (variable): A field containing Workstation data.
	Workstation []byte
}

// CreateNegotiateMessage initializes a NegotiateMessage with the given parameters
func CreateNegotiateMessage(domain, workstation string, negotiateFlags flags.NegotiateFlags, version *version.Version) (*NegotiateMessage, error) {
	msg := NegotiateMessage{
		Header: header.Header{
			Signature:   header.NTLM_SIGNATURE,
			MessageType: types.MESSAGE_TYPE_NEGOTIATE,
		},
		NegotiateFlags: negotiateFlags,
		Version:        version,
	}

	if version != nil && !msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_VERSION) {
		msg.NegotiateFlags |= flags.NTLMSSP_NEGOTIATE_VERSION
	}

	// DomainNameFields
	msg.DomainNameFields.Len = 0
	msg.DomainNameFields.MaxLen = 0
	msg.DomainNameFields.BufferOffset = 0
	if len(domain) > 0 {
		if !msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_OEM_DOMAIN_SUPPLIED) {
			msg.NegotiateFlags |= flags.NTLMSSP_NEGOTIATE_OEM_DOMAIN_SUPPLIED
		}
		if msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_UNICODE) {
			msg.DomainName = utf16.EncodeUTF16LE(strings.ToUpper(domain))
		} else {
			msg.DomainName = []byte(strings.ToUpper(domain))
		}
		msg.DomainNameFields.Len = uint16(len(msg.DomainName))
		msg.DomainNameFields.MaxLen = uint16(len(msg.DomainName))
	}

	// WorkstationFields
	msg.WorkstationFields.Len = 0
	msg.WorkstationFields.MaxLen = 0
	msg.WorkstationFields.BufferOffset = 0
	if len(workstation) > 0 {
		if !msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_OEM_WORKSTATION_SUPPLIED) {
			msg.NegotiateFlags |= flags.NTLMSSP_NEGOTIATE_OEM_WORKSTATION_SUPPLIED
		}
		if msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_UNICODE) {
			msg.Workstation = utf16.EncodeUTF16LE(strings.ToUpper(workstation))
		} else {
			msg.Workstation = []byte(strings.ToUpper(workstation))
		}
		msg.WorkstationFields.Len = uint16(len(msg.Workstation))
		msg.WorkstationFields.MaxLen = uint16(len(msg.Workstation))
	}

	return &msg, nil
}

// Marshal serializes the NegotiateMessage into a byte slice
func (msg *NegotiateMessage) Marshal() ([]byte, error) {
	// A 32-bit unsigned integer that defines the offset, in bytes, from
	// the beginning of the NEGOTIATE_MESSAGE to the entry in Payload
	// Compute the start‐of‐payload offset:
	//   8 bytes  NTLMSSP Signature
	// + 4 bytes  MessageType
	// + 4 bytes  NegotiateFlags
	// + 8 bytes  DomainNameFields
	// + 8 bytes  WorkstationFields
	// + 8 bytes  Version (real or zero placeholder)
	offset := 8 + 4 + 4 + 8 + 8 + 8 // = 40

	// Write payload data first to compute offsets
	payload := []byte{}

	// Domain name
	msg.DomainNameFields.BufferOffset = uint32(offset)
	if msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_OEM_DOMAIN_SUPPLIED) {
		payload = append(payload, msg.DomainName...)
		offset += len(msg.DomainName)
		msg.DomainNameFields.Len = uint16(len(msg.DomainName))
		msg.DomainNameFields.MaxLen = uint16(len(msg.DomainName))
	} else {
		msg.DomainNameFields.Len = 0
		msg.DomainNameFields.MaxLen = 0
	}

	// Workstation
	msg.WorkstationFields.BufferOffset = uint32(offset)
	if msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_OEM_WORKSTATION_SUPPLIED) {
		payload = append(payload, msg.Workstation...)
		offset += len(msg.Workstation)
		msg.WorkstationFields.Len = uint16(len(msg.Workstation))
		msg.WorkstationFields.MaxLen = uint16(len(msg.Workstation))
	} else {
		msg.WorkstationFields.Len = 0
		msg.WorkstationFields.MaxLen = 0
	}

	// Data section

	marshalledData := []byte{}

	// Create header section
	msg.Header.MessageType = types.MESSAGE_TYPE_NEGOTIATE
	msg.Header.Signature = header.NTLM_SIGNATURE
	marshalledHeader, err := msg.Header.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, marshalledHeader...)

	// Write negotiate flags
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(msg.NegotiateFlags))
	marshalledData = append(marshalledData, buf4...)

	// Write domain name fields
	domainNameFieldsBytes, err := msg.DomainNameFields.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, domainNameFieldsBytes...)

	// Write workstation fields
	workstationFieldsBytes, err := msg.WorkstationFields.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, workstationFieldsBytes...)

	// Write version if needed
	if msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_VERSION) {
		if msg.Version == nil {
			return nil, fmt.Errorf("NTLMSSP_NEGOTIATE_VERSION flag is set but Version is nil")
		}
		byteStream, err := msg.Version.Marshal()
		if err != nil {
			return nil, err
		}
		marshalledData = append(marshalledData, byteStream...)
	} else {
		marshalledData = append(marshalledData, []byte{0, 0, 0, 0, 0, 0, 0, 0}...)
	}

	// Write payload
	marshalledData = append(marshalledData, payload...)

	return marshalledData, nil
}

// Unmarshal deserializes a byte slice into a NegotiateMessage
func (msg *NegotiateMessage) Unmarshal(data []byte) (int, error) {
	totalBytesRead := 0

	if len(data) < 32 {
		return 0, fmt.Errorf("data too short to be a valid NegotiateMessage")
	}

	// Read header
	bytesRead, err := msg.Header.Unmarshal(data)
	if err != nil {
		return 0, err
	}
	if err := msg.Header.Expect(types.MESSAGE_TYPE_NEGOTIATE); err != nil {
		return 0, err
	}
	totalBytesRead += bytesRead

	// Read negotiate flags
	if totalBytesRead+4 > len(data) {
		return 0, fmt.Errorf("data too short to read NegotiateFlags in NegotiateMessage")
	}
	msg.NegotiateFlags = flags.NegotiateFlags(binary.LittleEndian.Uint32(data[totalBytesRead : totalBytesRead+4]))
	totalBytesRead += 4

	// Read domain fields
	if totalBytesRead+8 > len(data) {
		return 0, fmt.Errorf("data too short to read DomainNameFields in NegotiateMessage")
	}
	bytesRead, err = msg.DomainNameFields.Unmarshal(data[totalBytesRead : totalBytesRead+8])
	if err != nil {
		return 0, err
	}
	totalBytesRead += bytesRead

	// Read workstation fields
	if totalBytesRead+8 > len(data) {
		return 0, fmt.Errorf("data too short to read WorkstationFields in NegotiateMessage")
	}
	bytesRead, err = msg.WorkstationFields.Unmarshal(data[totalBytesRead : totalBytesRead+8])
	if err != nil {
		return 0, err
	}
	totalBytesRead += bytesRead

	// Read version if needed
	if (msg.NegotiateFlags & flags.NTLMSSP_NEGOTIATE_VERSION) != 0 {
		if totalBytesRead+8 > len(data) {
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

	// The payload bounds are computed in 64-bit. BufferOffset is a 32-bit field
	// taken straight off the wire, so adding Len to it in 32-bit arithmetic wraps:
	// an offset of 0xFFFFFFFF produces a sum that compares as inside the buffer,
	// and the slice expression below then panics. uint64 cannot overflow for a
	// uint32 offset plus a uint16 length, whatever the width of int on the host.

	// Domain name
	if msg.DomainNameFields.Len != 0 {
		domainStart := uint64(msg.DomainNameFields.BufferOffset)
		domainEnd := domainStart + uint64(msg.DomainNameFields.Len)
		if domainEnd > uint64(len(data)) {
			return 0, fmt.Errorf("data too short to read DomainName in payload section in NegotiateMessage")
		}
		msg.DomainName = data[domainStart:domainEnd]
	}

	// Workstation
	if msg.WorkstationFields.Len != 0 {
		workstationStart := uint64(msg.WorkstationFields.BufferOffset)
		workstationEnd := workstationStart + uint64(msg.WorkstationFields.Len)
		if workstationEnd > uint64(len(data)) {
			return 0, fmt.Errorf("data too short to read Workstation in payload section in NegotiateMessage")
		}
		msg.Workstation = data[workstationStart:workstationEnd]
	}

	return totalBytesRead, nil
}

// GetMessageType returns the message type of the NegotiateMessage
func (msg *NegotiateMessage) GetMessageType() uint32 {
	return uint32(msg.Header.MessageType)
}

// GetDomainName returns the domain name as a string
func (msg *NegotiateMessage) GetDomainName() string {
	if msg.NegotiateFlags&flags.NTLMSSP_NEGOTIATE_UNICODE != 0 {
		return utf16.DecodeUTF16LE(msg.DomainName)
	}
	return string(msg.DomainName)
}

// SetDomainName sets the domain name
func (msg *NegotiateMessage) SetDomainName(domain string) {
	if msg.NegotiateFlags&flags.NTLMSSP_NEGOTIATE_UNICODE != 0 {
		msg.DomainName = utf16.EncodeUTF16LE(domain)
	} else {
		msg.DomainName = []byte(strings.ToUpper(domain))
	}
	msg.DomainNameFields.Len = uint16(len(msg.DomainName))
	msg.DomainNameFields.MaxLen = uint16(len(msg.DomainName))
	if domain != "" {
		msg.NegotiateFlags |= flags.NTLMSSP_NEGOTIATE_OEM_DOMAIN_SUPPLIED
	}
}

// GetWorkstationName returns the workstation name as a string
func (msg *NegotiateMessage) GetWorkstationName() string {
	if msg.NegotiateFlags&flags.NTLMSSP_NEGOTIATE_UNICODE != 0 {
		return utf16.DecodeUTF16LE(msg.Workstation)
	}
	return string(msg.Workstation)
}

// SetWorkstationName sets the workstation name
func (msg *NegotiateMessage) SetWorkstationName(workstation string) {
	if msg.NegotiateFlags&flags.NTLMSSP_NEGOTIATE_UNICODE != 0 {
		msg.Workstation = utf16.EncodeUTF16LE(workstation)
	} else {
		msg.Workstation = []byte(strings.ToUpper(workstation))
	}
	msg.WorkstationFields.Len = uint16(len(msg.Workstation))
	msg.WorkstationFields.MaxLen = uint16(len(msg.Workstation))
	if workstation != "" {
		msg.NegotiateFlags |= flags.NTLMSSP_NEGOTIATE_OEM_WORKSTATION_SUPPLIED
	}
}
