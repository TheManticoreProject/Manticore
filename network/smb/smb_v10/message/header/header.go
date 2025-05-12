package header

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/securityfeatures"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
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
	Status types.ULONG
	// Flags (1 byte): An 8-bit field of 1-bit flags describing various features in effect for the message.
	Flags flags.Flags
	// Flags2 (2 bytes): A 16-bit field of 1-bit flags that represent various features in effect for the message.
	// Unspecified bits are reserved and MUST be zero.
	Flags2 flags.Flags2
	// PID (2 bytes): A 32-bit field that represents the process identifier (PID).
	PIDHigh types.USHORT
	// SecurityFeatures (8 bytes): This 8-byte field has three possible interpretations.
	SecurityFeatures securityfeatures.SecurityFeatures
	// Reserved (2 bytes): This field is reserved and SHOULD be set to 0x0000.
	Reserved types.USHORT
	// TID (2 bytes): A tree identifier (TID).
	TID types.USHORT
	// PIDLow (2 bytes): A 32-bit field that represents the process identifier (PID).
	PIDLow types.USHORT
	// UID (2 bytes): A user identifier (UID).
	UID types.USHORT
	// MID (2 bytes): A multiplex identifier (MID).
	MID types.USHORT
}

// NewHeader creates a new SMB Header with default values
// and initializes the SecurityFeatures field with a Reserved security features object.
//
// Returns:
//   - *Header: A pointer to the newly created SMB Header
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
//
// Returns:
//   - *Header: A pointer to the newly created SMB Header
func NewHeaderWithSecurityFeaturesConnectionLess() *Header {
	h := NewHeader()
	h.SecurityFeatures = securityfeatures.NewSecurityFeaturesConnectionlessTransport()
	return h
}

// NewHeaderWithSecurityFeaturesSecuritySignature creates a new SMB Header with default values
// and initializes the SecurityFeatures field with a SecuritySignature security features object.
//
// Returns:
//   - *Header: A pointer to the newly created SMB Header
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
	buf := []byte{}

	// Protocol (4 bytes)
	buf = append(buf, h.Protocol[0:4]...)

	// Command (1 byte)
	buf = append(buf, byte(h.Command))

	// Status (4 bytes)
	buf4 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf4, h.Status)
	buf = append(buf, buf4...)

	// Flags (1 byte)
	buf = append(buf, byte(h.Flags))

	// Flags2 (2 bytes)
	buf2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, uint16(h.Flags2))
	buf = append(buf, buf2...)

	// PIDHigh (2 bytes)
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, h.PIDHigh)
	buf = append(buf, buf2...)

	// SecurityFeatures (8 bytes)
	securityFeaturesBytes, err := h.SecurityFeatures.Marshal()
	if err != nil {
		return nil, err
	}
	buf = append(buf, securityFeaturesBytes...)

	// Reserved (2 bytes)
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, h.Reserved)
	buf = append(buf, buf2...)

	// TID (2 bytes)
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, h.TID)
	buf = append(buf, buf2...)

	// PIDLow (2 bytes)
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, h.PIDLow)
	buf = append(buf, buf2...)

	// UID (2 bytes)
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, h.UID)
	buf = append(buf, buf2...)

	// MID (2 bytes)
	buf2 = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf2, h.MID)
	buf = append(buf, buf2...)

	if len(buf) != SMB_HEADER_SIZE {
		return nil, fmt.Errorf("expected to have marshalled %d bytes, got %d", SMB_HEADER_SIZE, len(buf))
	}

	return buf, nil
}

// Unmarshal deserializes a byte slice into the SMB Header structure.
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
func (h *Header) Unmarshal(data []byte) (int, error) {
	if len(data) < SMB_HEADER_SIZE {
		return 0, fmt.Errorf("data too short to unmarshal SMB header")
	}

	bytesRead := 0

	// Protocol (4 bytes)
	copy(h.Protocol[:], data[0:4])
	bytesRead += 4

	// Command (1 byte)
	h.Command = codes.CommandCode(data[bytesRead])
	bytesRead += 1

	// Status (4 bytes)
	h.Status = binary.LittleEndian.Uint32(data[bytesRead : bytesRead+4])
	bytesRead += 4

	// Flags (1 byte)
	h.Flags = flags.Flags(data[bytesRead])
	bytesRead += 1

	// Flags2 (2 bytes)
	h.Flags2 = flags.Flags2(binary.LittleEndian.Uint16(data[bytesRead : bytesRead+2]))
	bytesRead += 2

	// PIDHigh (2 bytes)
	h.PIDHigh = binary.LittleEndian.Uint16(data[bytesRead : bytesRead+2])
	bytesRead += 2

	// SecurityFeatures (8 bytes)
	securityFeaturesBytes := data[bytesRead : bytesRead+8]
	h.SecurityFeatures = securityfeatures.NewSecurityFeaturesReserved()
	securityFeaturesBytesRead, err := h.SecurityFeatures.Unmarshal(securityFeaturesBytes)
	if err != nil {
		return 0, err
	}
	if securityFeaturesBytesRead != 8 {
		return 0, fmt.Errorf("expected to have unmarshalled 8 bytes, got %d", securityFeaturesBytesRead)
	}
	bytesRead += securityFeaturesBytesRead

	// Reserved (2 bytes)
	h.Reserved = binary.LittleEndian.Uint16(data[bytesRead : bytesRead+2])
	bytesRead += 2

	// TID (2 bytes)
	h.TID = binary.LittleEndian.Uint16(data[bytesRead : bytesRead+2])
	bytesRead += 2

	// PIDLow (2 bytes)
	h.PIDLow = binary.LittleEndian.Uint16(data[bytesRead : bytesRead+2])
	bytesRead += 2

	// UID (2 bytes)
	h.UID = binary.LittleEndian.Uint16(data[bytesRead : bytesRead+2])
	bytesRead += 2

	// MID (2 bytes)
	h.MID = binary.LittleEndian.Uint16(data[bytesRead : bytesRead+2])
	bytesRead += 2

	if bytesRead != SMB_HEADER_SIZE {
		return 0, fmt.Errorf("expected to have unmarshalled %d bytes, got %d", SMB_HEADER_SIZE, bytesRead)
	}

	return bytesRead, nil
}

// GetMID returns the multiplex identifier (MID) value
//
// Returns:
//   - types.USHORT: The multiplex identifier (MID) value
func (h *Header) GetMID() types.USHORT {
	return h.MID
}

// SetMID sets the multiplex identifier (MID) value
//
// Parameters:
//   - mid: The new multiplex identifier (MID) value
func (h *Header) SetMID(mid types.USHORT) {
	h.MID = mid
}

// GetPID returns the full 32-bit process identifier (PID) value
// combining PIDHigh and PIDLow
//
// Returns:
//   - types.ULONG: The full 32-bit process identifier (PID) value
func (h *Header) GetPID() types.ULONG {
	return (types.ULONG(h.PIDHigh) << 16) | types.ULONG(h.PIDLow)
}

// SetPID sets both PIDHigh and PIDLow from a 32-bit process identifier (PID)
//
// Parameters:
//   - pid: The new 32-bit process identifier (PID) value
func (h *Header) SetPID(pid types.ULONG) {
	h.PIDHigh = types.USHORT(pid >> 16)
	h.PIDLow = types.USHORT(pid & 0xFFFF)
}

// GetTID returns the tree identifier (TID) value
//
// Returns:
//   - types.USHORT: The tree identifier (TID) value
func (h *Header) GetTID() types.USHORT {
	return h.TID
}

// SetTID sets the tree identifier (TID) value
//
// Parameters:
//   - tid: The new tree identifier (TID) value
func (h *Header) SetTID(tid types.USHORT) {
	h.TID = tid
}

// GetUID returns the user identifier (UID) value
//
// Returns:
//   - types.USHORT: The user identifier (UID) value
func (h *Header) GetUID() types.USHORT {
	return h.UID
}

// SetUID sets the user identifier (UID) value
//
// Parameters:
//   - uid: The new user identifier (UID) value
func (h *Header) SetUID(uid types.USHORT) {
	h.UID = uid
}
