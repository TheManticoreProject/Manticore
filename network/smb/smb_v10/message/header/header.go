package header

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/securityfeatures"
)

const (
	SMB_HEADER_SIZE = 32
)

// Header represents the header structure for SMB packets
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/69a29f73-de0c-45a6-a1aa-8ceeea42217f
type Header struct {
	// Protocol (4 bytes): This field MUST contain the 4-byte literal string '\xFF', 'S', 'M', 'B',
	// with the letters represented by their respective ASCII values in the order shown. In the earliest
	// available SMB documentation, this field is defined as a one byte message type (0xFF) followed by
	// a three byte server type identifier.
	Protocol [4]byte
	// Command (1 byte): A one-byte command code. Defined SMB command codes are listed in section 2.2.2.1.
	Command codes.CommandCode
	// Status (4 bytes): A 32-bit field used to communicate error messages from the server to the client.
	Status uint32
	// Flags (1 byte): An 8-bit field of 1-bit flags describing various features in effect for the message.
	Flags byte
	// Flags2 (2 bytes): A 16-bit field of 1-bit flags that represent various features in effect for the message.
	// Unspecified bits are reserved and MUST be zero.
	Flags2 uint16
	// PIDHigh (2 bytes): If set to a nonzero value, this field represents the high-order bytes of a process
	// identifier (PID). It is combined with the PIDLow field below to form a full PID.
	PIDHigh uint16
	// SecurityFeatures (8 bytes): This 8-byte field has three possible interpretations.
	SecurityFeatures securityfeatures.SecurityFeatures
	// Reserved (2 bytes): This field is reserved and SHOULD be set to 0x0000.
	Reserved uint16
	// TID (2 bytes): A tree identifier (TID).
	TID uint16
	// PIDLow (2 bytes): The lower 16-bits of the PID.
	PIDLow uint16
	// UID (2 bytes): A user identifier (UID).
	UID uint16
	// MID (2 bytes): A multiplex identifier (MID).
	MID uint16
}

// NewHeader creates a new SMB Header with default values
// and initializes the SecurityFeatures field with a Reserved security features object.
func NewHeader() *Header {
	h := &Header{
		Protocol:         [4]byte{0xFF, 'S', 'M', 'B'},
		Command:          0x00,
		Status:           0x00000000,
		Flags:            0x00,
		Flags2:           0x0000,
		PIDHigh:          0x0000,
		SecurityFeatures: securityfeatures.NewSecurityFeaturesReserved(),
		Reserved:         0x0000,
		TID:              0x0000,
		PIDLow:           0x0000,
		UID:              0x0000,
		MID:              0x0000,
	}

	return h
}

// NewHeaderWithSecurityFeaturesConnectionLess creates a new SMB Header with default values
// and initializes the SecurityFeatures field with a ConnectionlessTransport security features object.
func NewHeaderWithSecurityFeaturesConnectionLess() *Header {
	h := NewHeader()
	h.SecurityFeatures = securityfeatures.NewSecurityFeaturesConnectionlessTransport()
	return h
}

// NewHeaderWithSecurityFeaturesSecuritySignature creates a new SMB Header with default values
// and initializes the SecurityFeatures field with a SecuritySignature security features object.
func NewHeaderWithSecurityFeaturesSecuritySignature() *Header {
	h := NewHeader()
	h.SecurityFeatures = securityfeatures.NewSecurityFeaturesSecuritySignature()
	return h
}

// Marshal serializes the SMB Header structure into a byte slice.
// It converts all header fields into their binary representation using little-endian byte order.
// The resulting byte slice will be exactly 32 bytes (SMB_HEADER_SIZE) long, containing
// all header fields in the order specified by the SMB protocol.
//
// Returns:
//   - []byte: The serialized header as a byte slice
//   - error: Any error encountered during serialization, or nil if successful
func (h *Header) Marshal() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})

	binary.Write(buf, binary.LittleEndian, h.Protocol)

	binary.Write(buf, binary.LittleEndian, h.Command)

	binary.Write(buf, binary.LittleEndian, h.Status)

	binary.Write(buf, binary.LittleEndian, h.Flags)

	binary.Write(buf, binary.LittleEndian, h.Flags2)

	binary.Write(buf, binary.LittleEndian, h.PIDHigh)

	securityFeaturesBytes, err := h.SecurityFeatures.Marshal()
	if err != nil {
		return nil, err
	}
	buf.Write(securityFeaturesBytes)

	binary.Write(buf, binary.LittleEndian, h.Reserved)

	binary.Write(buf, binary.LittleEndian, h.TID)

	binary.Write(buf, binary.LittleEndian, h.PIDLow)

	binary.Write(buf, binary.LittleEndian, h.UID)

	binary.Write(buf, binary.LittleEndian, h.MID)

	return buf.Bytes(), nil
}

// UnmarshalWithSecurityFeaturesConnectionlessTransport deserializes a byte slice into the SMB Header structure.
// It reads the binary representation of the header fields from the input byte slice
// using little-endian byte order. This method determines the appropriate security
// features type based on the header flags and unmarshals accordingly.
//
// Parameters:
//   - data: The byte slice containing the serialized SMB header
//
// Returns:
//   - int: The number of bytes read from the input byte slice
//   - error: Any error encountered during deserialization, or nil if successful
func (h *Header) UnmarshalWithSecurityFeaturesConnectionlessTransport(data []byte) (int, error) {
	if len(data) < SMB_HEADER_SIZE {
		return 0, fmt.Errorf("data too short to unmarshal SMB header")
	}

	copy(h.Protocol[:], data[0:4])

	h.Command = codes.CommandCode(data[4])

	h.Status = binary.LittleEndian.Uint32(data[5:9])

	h.Flags = data[9]

	h.Flags2 = binary.LittleEndian.Uint16(data[10:12])

	h.PIDHigh = binary.LittleEndian.Uint16(data[12:14])

	securityFeaturesBytes := data[14:22]
	h.SecurityFeatures = securityfeatures.NewSecurityFeaturesConnectionlessTransport()
	bytesRead, err := h.SecurityFeatures.Unmarshal(securityFeaturesBytes)
	if err != nil {
		return 0, err
	}
	if bytesRead != 8 {
		return 0, fmt.Errorf("expected 8 bytes, got %d", bytesRead)
	}

	h.Reserved = binary.LittleEndian.Uint16(data[22:24])

	h.TID = binary.LittleEndian.Uint16(data[24:26])

	h.PIDLow = binary.LittleEndian.Uint16(data[26:28])

	h.UID = binary.LittleEndian.Uint16(data[28:30])

	h.MID = binary.LittleEndian.Uint16(data[30:32])

	return SMB_HEADER_SIZE, nil
}

// UnmarshalWithSecurityFeaturesSecuritySignature deserializes a byte slice into the SMB Header structure.
// It reads the binary representation of the header fields from the input byte slice
// using little-endian byte order. This method determines the appropriate security
// features type based on the header flags and unmarshals accordingly.
//
// Parameters:
//   - data: The byte slice containing the serialized SMB header
//
// Returns:
//   - int: The number of bytes read from the input byte slice
//   - error: Any error encountered during deserialization, or nil if successful
func (h *Header) UnmarshalWithSecurityFeaturesSecuritySignature(data []byte) (int, error) {
	if len(data) < SMB_HEADER_SIZE {
		return 0, fmt.Errorf("data too short to unmarshal SMB header")
	}

	copy(h.Protocol[:], data[0:4])

	h.Command = codes.CommandCode(data[4])

	h.Status = binary.LittleEndian.Uint32(data[5:9])

	h.Flags = data[9]

	h.Flags2 = binary.LittleEndian.Uint16(data[10:12])

	h.PIDHigh = binary.LittleEndian.Uint16(data[12:14])

	securityFeaturesBytes := data[14:22]
	h.SecurityFeatures = securityfeatures.NewSecurityFeaturesSecuritySignature()
	bytesRead, err := h.SecurityFeatures.Unmarshal(securityFeaturesBytes)
	if err != nil {
		return 0, err
	}
	if bytesRead != 8 {
		return 0, fmt.Errorf("expected 8 bytes, got %d", bytesRead)
	}

	h.Reserved = binary.LittleEndian.Uint16(data[22:24])

	h.TID = binary.LittleEndian.Uint16(data[24:26])

	h.PIDLow = binary.LittleEndian.Uint16(data[26:28])

	h.UID = binary.LittleEndian.Uint16(data[28:30])

	h.MID = binary.LittleEndian.Uint16(data[30:32])

	return SMB_HEADER_SIZE, nil
}

// UnmarshalWithSecurityFeaturesConnectionlessTransport deserializes a byte slice into the SMB Header structure.
// It reads the binary representation of the header fields from the input byte slice
// using little-endian byte order. This method determines the appropriate security
// features type based on the header flags and unmarshals accordingly.
//
// Parameters:
//   - data: The byte slice containing the serialized SMB header
//
// Returns:
//   - int: The number of bytes read from the input byte slice
//   - error: Any error encountered during deserialization, or nil if successful
func (h *Header) UnmarshalWithSecurityFeaturesReserved(data []byte) (int, error) {
	if len(data) < SMB_HEADER_SIZE {
		return 0, fmt.Errorf("data too short to unmarshal SMB header")
	}

	copy(h.Protocol[:], data[0:4])

	h.Command = codes.CommandCode(data[4])

	h.Status = binary.LittleEndian.Uint32(data[5:9])

	h.Flags = data[9]

	h.Flags2 = binary.LittleEndian.Uint16(data[10:12])

	h.PIDHigh = binary.LittleEndian.Uint16(data[12:14])

	securityFeaturesBytes := data[14:22]
	h.SecurityFeatures = securityfeatures.NewSecurityFeaturesReserved()
	bytesRead, err := h.SecurityFeatures.Unmarshal(securityFeaturesBytes)
	if err != nil {
		return 0, err
	}

	if bytesRead != 8 {
		return 0, fmt.Errorf("expected 8 bytes, got %d", bytesRead)
	}

	h.Reserved = binary.LittleEndian.Uint16(data[22:24])

	h.TID = binary.LittleEndian.Uint16(data[24:26])

	h.PIDLow = binary.LittleEndian.Uint16(data[26:28])

	h.UID = binary.LittleEndian.Uint16(data[28:30])

	h.MID = binary.LittleEndian.Uint16(data[30:32])

	return SMB_HEADER_SIZE, nil
}

func (h *Header) Unmarshal(data []byte) (int, error) {
	return h.UnmarshalWithSecurityFeaturesReserved(data)
}
