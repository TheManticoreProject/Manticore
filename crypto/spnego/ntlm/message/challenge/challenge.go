package challenge

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/datafields"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/header"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/types"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/version"
)

// ChallengeMessage is the second message in NTLM authentication
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/801a4681-8809-4be9-ab0d-61dcfe762786
type ChallengeMessage struct {
	header.Header

	// TargetNameFields (8 bytes): A field containing TargetName information.
	TargetNameFields datafields.DataFields

	// NegotiateFlags (4 bytes): A NEGOTIATE structure that contains a set of flags, as defined by section 2.2.2.5. The server sets flags to indicate options it supports or, if there has been a NEGOTIATE_MESSAGE (section 2.2.1.1), the choices it has made from the options offered by the client. If the client has set the NTLMSSP_NEGOTIATE_SIGN in the NEGOTIATE_MESSAGE the Server MUST return it.
	NegotiateFlags flags.NegotiateFlags

	// ServerChallenge (8 bytes): A 64-bit value that contains the NTLM challenge. The challenge is a 64-bit nonce. The processing of the ServerChallenge is specified in sections 3.1.5 and 3.2.5.
	ServerChallenge [8]byte

	// Reserved (8 bytes): An 8-byte array whose elements MUST be zero when sent and MUST be ignored on receipt.
	Reserved [8]byte

	// TargetInfoFields (8 bytes): A field containing TargetInfo information.
	TargetInfoFields datafields.DataFields

	// Version (8 bytes): A VERSION structure that contains version information.
	Version *version.Version

	// Payload section

	// TargetName (variable): A field containing TargetName data.
	TargetName []byte

	// TargetInfo (variable): A field containing TargetInfo data.
	TargetInfo []byte
}

// Marshal serializes the ChallengeMessage into a byte slice
func (msg *ChallengeMessage) Marshal() ([]byte, error) {
	// A 32-bit unsigned integer that defines the offset, in bytes, from
	// the beginning of the NEGOTIATE_MESSAGE to the entry in Payload
	// Starting at 56 for the header section (8+4 bytes) + 8 for target name fields + 4 for negotiate flags + 8 for server challenge + 8 for reserved + 8 for target info fields + 8 for version
	offset := 56 // (8 + 4 + 8 + 4 + 8 + 8 + 8 + 8)

	// Write payload data first to compute offsets
	payload := []byte{}

	// Target name
	msg.TargetNameFields.Len = uint16(len(msg.TargetName))
	msg.TargetNameFields.MaxLen = uint16(len(msg.TargetName))
	msg.TargetNameFields.BufferOffset = uint32(offset)
	offset += len(msg.TargetName)
	payload = append(payload, msg.TargetName...)

	// Target info
	msg.TargetInfoFields.Len = uint16(len(msg.TargetInfo))
	msg.TargetInfoFields.MaxLen = uint16(len(msg.TargetInfo))
	msg.TargetInfoFields.BufferOffset = uint32(offset)
	offset += len(msg.TargetInfo)
	payload = append(payload, msg.TargetInfo...)

	// Data section

	marshalledData := []byte{}

	// Create header section
	msg.Header.MessageType = types.MESSAGE_TYPE_CHALLENGE
	msg.Header.Signature = header.NTLM_SIGNATURE
	marshalledHeader, err := msg.Header.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, marshalledHeader...)

	// Write target name fields
	targetNameFieldsBytes, err := msg.TargetNameFields.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, targetNameFieldsBytes...)

	// Write negotiate flags
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, uint32(msg.NegotiateFlags))
	marshalledData = append(marshalledData, buf4...)

	// Write server challenge (8 bytes)
	marshalledData = append(marshalledData, msg.ServerChallenge[:]...)

	// Write reserved (8 bytes)
	marshalledData = append(marshalledData, msg.Reserved[:]...)

	// Write target info fields
	targetInfoFieldsBytes, err := msg.TargetInfoFields.Marshal()
	if err != nil {
		return nil, err
	}
	marshalledData = append(marshalledData, targetInfoFieldsBytes...)

	// Write version if needed
	if msg.Version != nil && msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_VERSION) {
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

// Unmarshal deserializes the ChallengeMessage from a byte slice
func (msg *ChallengeMessage) Unmarshal(data []byte) (int, error) {
	totalBytesRead := 0

	if len(data) < 60 {
		return 0, fmt.Errorf("data too short to be a valid ChallengeMessage")
	}

	// Read header
	bytesRead, err := msg.Header.Unmarshal(data)
	if err != nil {
		return 0, err
	}
	if err := msg.Header.Expect(types.MESSAGE_TYPE_CHALLENGE); err != nil {
		return 0, err
	}
	totalBytesRead += bytesRead

	// Read target name fields
	if len(data) < 16 {
		return 0, fmt.Errorf("data too short to read TargetNameFields in ChallengeMessage")
	}
	bytesRead, err = msg.TargetNameFields.Unmarshal(data[totalBytesRead:])
	if err != nil {
		return 0, err
	}
	totalBytesRead += bytesRead

	// Read negotiate flags
	if len(data[totalBytesRead:]) < 4 {
		return 0, fmt.Errorf("data too short to read NegotiateFlags in ChallengeMessage")
	}
	msg.NegotiateFlags = flags.NegotiateFlags(binary.LittleEndian.Uint32(data[totalBytesRead:]))
	totalBytesRead += 4

	// Read server challenge
	if len(data[totalBytesRead:]) < 8 {
		return 0, fmt.Errorf("data too short to read ServerChallenge in ChallengeMessage")
	}
	copy(msg.ServerChallenge[:], data[totalBytesRead:totalBytesRead+8])
	totalBytesRead += 8

	// Read reserved
	if len(data[totalBytesRead:]) < 8 {
		return 0, fmt.Errorf("data too short to read Reserved in ChallengeMessage")
	}
	copy(msg.Reserved[:], data[totalBytesRead:totalBytesRead+8])
	totalBytesRead += 8

	// Read target info fields
	if len(data[totalBytesRead:]) < 8 {
		return 0, fmt.Errorf("data too short to read TargetInfoFields in ChallengeMessage")
	}
	bytesRead, err = msg.TargetInfoFields.Unmarshal(data[totalBytesRead:])
	if err != nil {
		return 0, err
	}
	totalBytesRead += bytesRead

	// Read version if needed
	if (msg.NegotiateFlags & flags.NTLMSSP_NEGOTIATE_VERSION) != 0 {
		if totalBytesRead+8 > len(data) {
			return 0, fmt.Errorf("data too short to read Version in ChallengeMessage")
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

	// Target name
	targetNameStart := uint64(msg.TargetNameFields.BufferOffset)
	targetNameEnd := targetNameStart + uint64(msg.TargetNameFields.Len)
	if targetNameEnd > uint64(len(data)) {
		return 0, fmt.Errorf("data too short to read TargetName in payload section in ChallengeMessage")
	}
	msg.TargetName = data[targetNameStart:targetNameEnd]

	// Target info
	targetInfoStart := uint64(msg.TargetInfoFields.BufferOffset)
	targetInfoEnd := targetInfoStart + uint64(msg.TargetInfoFields.Len)
	if targetInfoEnd > uint64(len(data)) {
		return 0, fmt.Errorf("data too short to read TargetInfo in payload section in ChallengeMessage")
	}
	msg.TargetInfo = data[targetInfoStart:targetInfoEnd]

	return totalBytesRead, nil
}

// GetMessageType returns the message type of the ChallengeMessage
func (msg *ChallengeMessage) GetMessageType() uint32 {
	return uint32(msg.Header.MessageType)
}
