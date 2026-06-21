package regf

import (
	"encoding/binary"
	"fmt"
)

const (
	skSignature = 0x6B73 // "sk" in little-endian
	skMinSize   = 20     // header before the embedded security descriptor
)

// SecurityKey is a parsed SK (key security) record. Multiple key nodes can share one SK
// record (it is reference-counted); a KeyNode reaches its SK record through SecurityOffset.
//
// Source: https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md#key-security
type SecurityKey struct {
	// Signature (2 bytes): must be ASCII "sk" (0x6B73 little-endian).
	Signature uint16

	// Reserved (2 bytes): unused.
	Reserved uint16

	// Flink (4 bytes): offset of the next sk record in the circular list.
	Flink uint32

	// Blink (4 bytes): offset of the previous sk record in the circular list.
	Blink uint32

	// ReferenceCount (4 bytes): number of key nodes referencing this record.
	ReferenceCount uint32

	// SecurityDescriptorSize (4 bytes): size in bytes of the embedded security descriptor.
	SecurityDescriptorSize uint32

	// SecurityDescriptor (variable): self-relative SECURITY_DESCRIPTOR blob ([MS-DTYP] 2.4.6).
	SecurityDescriptor []byte
}

// NewSecurityKey creates a new empty SecurityKey.
func NewSecurityKey() *SecurityKey {
	return &SecurityKey{}
}

// Unmarshal deserializes a SecurityKey from cell data (after the 4-byte cell size prefix).
//
// Parameters:
//   - data ([]byte): cell data starting with the "sk" signature.
//
// Returns:
//   - The number of bytes consumed.
//   - An error if the data is too short or the signature is invalid.
func (s *SecurityKey) Unmarshal(data []byte) (int, error) {
	if len(data) < skMinSize {
		return 0, fmt.Errorf("data too short for SecurityKey: need %d bytes, got %d", skMinSize, len(data))
	}

	s.Signature = binary.LittleEndian.Uint16(data[0:2])
	if s.Signature != skSignature {
		return 0, fmt.Errorf("invalid SecurityKey Signature: 0x%04X (expected 0x%04X)", s.Signature, skSignature)
	}

	s.Reserved = binary.LittleEndian.Uint16(data[2:4])
	s.Flink = binary.LittleEndian.Uint32(data[4:8])
	s.Blink = binary.LittleEndian.Uint32(data[8:12])
	s.ReferenceCount = binary.LittleEndian.Uint32(data[12:16])
	s.SecurityDescriptorSize = binary.LittleEndian.Uint32(data[16:20])

	sdEnd := skMinSize + int(s.SecurityDescriptorSize)
	if sdEnd > len(data) {
		sdEnd = len(data)
	}
	s.SecurityDescriptor = make([]byte, sdEnd-skMinSize)
	copy(s.SecurityDescriptor, data[skMinSize:sdEnd])

	return sdEnd, nil
}

// Marshal serializes the SecurityKey to binary data.
//
// Returns:
//   - A byte slice containing the serialized SecurityKey.
//   - An error if serialization fails.
func (s *SecurityKey) Marshal() ([]byte, error) {
	buf := make([]byte, skMinSize+len(s.SecurityDescriptor))

	binary.LittleEndian.PutUint16(buf[0:2], s.Signature)
	binary.LittleEndian.PutUint16(buf[2:4], s.Reserved)
	binary.LittleEndian.PutUint32(buf[4:8], s.Flink)
	binary.LittleEndian.PutUint32(buf[8:12], s.Blink)
	binary.LittleEndian.PutUint32(buf[12:16], s.ReferenceCount)
	binary.LittleEndian.PutUint32(buf[16:20], s.SecurityDescriptorSize)
	copy(buf[skMinSize:], s.SecurityDescriptor)

	return buf, nil
}
