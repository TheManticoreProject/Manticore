package regf

import (
	"encoding/binary"
	"fmt"
)

const (
	lfSig = 0x666C // "lf"
	lhSig = 0x686C // "lh"
	riSig = 0x6972 // "ri"
	liSig = 0x696C // "li"
)

// SubKeyList is the internal representation of a parsed subkey list record (LF, LH, LI, or RI).
//
// Source: https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md#subkeys-list
type SubKeyList struct {
	// Signature (2 bytes): "lf", "lh", "ri", or "li".
	Signature uint16

	// NumberOfElements (2 bytes): count of entries.
	NumberOfElements uint16

	// Elements contains the raw element data, interpreted per signature type.
	Elements []byte
}

// NewSubKeyList creates a new empty SubKeyList.
func NewSubKeyList() *SubKeyList {
	return &SubKeyList{}
}

// Unmarshal deserializes a SubKeyList from cell data.
//
// Parameters:
//   - data ([]byte): cell data starting with the list signature.
//
// Returns:
//   - The number of bytes consumed.
//   - An error if the data is too short or the signature is unrecognized.
func (s *SubKeyList) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short for SubKeyList: need 4 bytes, got %d", len(data))
	}

	s.Signature = binary.LittleEndian.Uint16(data[0:2])
	s.NumberOfElements = binary.LittleEndian.Uint16(data[2:4])

	var elemSize int
	switch s.Signature {
	case lfSig, lhSig:
		elemSize = 8 // 4-byte offset + 4-byte name hint/hash
	case riSig, liSig:
		elemSize = 4 // 4-byte offset only
	default:
		return 0, fmt.Errorf("unrecognized SubKeyList Signature: 0x%04X", s.Signature)
	}

	totalSize := 4 + int(s.NumberOfElements)*elemSize
	if totalSize > len(data) {
		totalSize = len(data)
	}
	s.Elements = make([]byte, totalSize-4)
	copy(s.Elements, data[4:totalSize])

	return totalSize, nil
}

// Marshal serializes the SubKeyList to binary data.
//
// Returns:
//   - A byte slice containing the serialized SubKeyList.
//   - An error if serialization fails.
func (s *SubKeyList) Marshal() ([]byte, error) {
	buf := make([]byte, 4+len(s.Elements))
	binary.LittleEndian.PutUint16(buf[0:2], s.Signature)
	binary.LittleEndian.PutUint16(buf[2:4], s.NumberOfElements)
	copy(buf[4:], s.Elements)
	return buf, nil
}

// KeyNodeOffsets returns the offsets to all key nodes referenced by this list.
// For RI lists, it returns the offsets to the sublists (not the key nodes themselves);
// the caller must recurse into each sublist.
func (s *SubKeyList) KeyNodeOffsets() []uint32 {
	n := int(s.NumberOfElements)
	offsets := make([]uint32, 0, n)

	switch s.Signature {
	case lfSig, lhSig:
		for i := 0; i < n; i++ {
			off := i * 8
			if off+4 > len(s.Elements) {
				break
			}
			offsets = append(offsets, binary.LittleEndian.Uint32(s.Elements[off:off+4]))
		}
	case riSig, liSig:
		for i := 0; i < n; i++ {
			off := i * 4
			if off+4 > len(s.Elements) {
				break
			}
			offsets = append(offsets, binary.LittleEndian.Uint32(s.Elements[off:off+4]))
		}
	}

	return offsets
}

// IsIndexRoot reports whether this is an RI (index root) list, which contains
// references to other sublists rather than directly to key nodes.
func (s *SubKeyList) IsIndexRoot() bool {
	return s.Signature == riSig
}
