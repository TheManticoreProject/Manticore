package regf

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	vkSignature     = 0x6B76 // "vk" in little-endian
	keyValueMinSize = 20
)

// Value data types.
//
// Source: https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md#key-value
const (
	RegNone                     uint32 = 0x00000000
	RegSz                       uint32 = 0x00000001
	RegExpandSz                 uint32 = 0x00000002
	RegBinary                   uint32 = 0x00000003
	RegDword                    uint32 = 0x00000004
	RegDwordBigEndian           uint32 = 0x00000005
	RegLink                     uint32 = 0x00000006
	RegMultiSz                  uint32 = 0x00000007
	RegResourceList             uint32 = 0x00000008
	RegFullResourceDescriptor   uint32 = 0x00000009
	RegResourceRequirementsList uint32 = 0x0000000A
	RegQword                    uint32 = 0x0000000B
)

// Value flags.
const (
	ValueCompName uint16 = 0x0001
)

// KeyValue is a parsed VK (value key) record representing a registry value.
//
// Source: https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md#key-value
type KeyValue struct {
	// Signature (2 bytes): must be ASCII "vk" (0x6B76 little-endian).
	Signature uint16

	// NameLength (2 bytes): length of value name in bytes; 0 = default (unnamed) value.
	NameLength uint16

	// DataSize (4 bytes): data size. Bit 31 set means inline data.
	DataSize uint32

	// DataOffset (4 bytes): offset to data cell, or inline data when bit 31 of DataSize is set.
	DataOffset uint32

	// DataType (4 bytes): REG_* data type constant.
	DataType uint32

	// Flags (2 bytes): VALUE_COMP_NAME etc.
	Flags uint16

	// Spare (2 bytes): unused.
	Spare uint16

	// NameRaw (variable): raw value name bytes.
	NameRaw []byte

	// hive back-reference for data retrieval.
	hive *Hive
}

// NewKeyValue creates a new empty KeyValue.
func NewKeyValue() *KeyValue {
	return &KeyValue{}
}

// Unmarshal deserializes a KeyValue from cell data.
//
// Parameters:
//   - data ([]byte): cell data starting with the "vk" signature.
//
// Returns:
//   - The number of bytes consumed.
//   - An error if the data is too short or the signature is invalid.
func (v *KeyValue) Unmarshal(data []byte) (int, error) {
	if len(data) < keyValueMinSize {
		return 0, fmt.Errorf("data too short for KeyValue: need %d bytes, got %d", keyValueMinSize, len(data))
	}

	v.Signature = binary.LittleEndian.Uint16(data[0:2])
	if v.Signature != vkSignature {
		return 0, fmt.Errorf("invalid KeyValue Signature: 0x%04X (expected 0x%04X)", v.Signature, vkSignature)
	}

	v.NameLength = binary.LittleEndian.Uint16(data[2:4])
	v.DataSize = binary.LittleEndian.Uint32(data[4:8])
	v.DataOffset = binary.LittleEndian.Uint32(data[8:12])
	v.DataType = binary.LittleEndian.Uint32(data[12:16])
	v.Flags = binary.LittleEndian.Uint16(data[16:18])
	v.Spare = binary.LittleEndian.Uint16(data[18:20])

	nameEnd := keyValueMinSize + int(v.NameLength)
	if nameEnd > len(data) {
		nameEnd = len(data)
	}
	if v.NameLength > 0 {
		v.NameRaw = make([]byte, nameEnd-keyValueMinSize)
		copy(v.NameRaw, data[keyValueMinSize:nameEnd])
	}

	return nameEnd, nil
}

// Marshal serializes the KeyValue to binary data.
//
// Returns:
//   - A byte slice containing the serialized KeyValue.
//   - An error if serialization fails.
func (v *KeyValue) Marshal() ([]byte, error) {
	size := keyValueMinSize + len(v.NameRaw)
	buf := make([]byte, size)

	binary.LittleEndian.PutUint16(buf[0:2], v.Signature)
	binary.LittleEndian.PutUint16(buf[2:4], v.NameLength)
	binary.LittleEndian.PutUint32(buf[4:8], v.DataSize)
	binary.LittleEndian.PutUint32(buf[8:12], v.DataOffset)
	binary.LittleEndian.PutUint32(buf[12:16], v.DataType)
	binary.LittleEndian.PutUint16(buf[16:18], v.Flags)
	binary.LittleEndian.PutUint16(buf[18:20], v.Spare)
	copy(buf[keyValueMinSize:], v.NameRaw)

	return buf, nil
}

// Name returns the decoded value name as a Go string. Returns "" for the default value.
func (v *KeyValue) Name() string {
	if v.NameLength == 0 {
		return ""
	}
	if v.Flags&ValueCompName != 0 {
		return string(v.NameRaw)
	}
	return decodeUTF16LE(v.NameRaw)
}

// Type returns the REG_* data type.
func (v *KeyValue) Type() uint32 {
	return v.DataType
}

// IsInline reports whether the value data is stored inline in the DataOffset field.
func (v *KeyValue) IsInline() bool {
	return v.DataSize&0x80000000 != 0
}

// ActualDataSize returns the data size with the inline flag masked off.
func (v *KeyValue) ActualDataSize() uint32 {
	return v.DataSize & 0x7FFFFFFF
}

// Data returns the raw data bytes for this value.
func (v *KeyValue) Data() ([]byte, error) {
	actualSize := v.ActualDataSize()
	if actualSize == 0 {
		return nil, nil
	}

	if v.IsInline() {
		// Data stored in the DataOffset field (up to 4 bytes)
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, v.DataOffset)
		if actualSize > 4 {
			actualSize = 4
		}
		return buf[:actualSize], nil
	}

	if v.hive == nil {
		return nil, fmt.Errorf("KeyValue not attached to a hive")
	}

	// Values larger than a single data cell are stored in a big-data (db) record that
	// chains multiple data segments; reassemble them transparently.
	if actualSize > bigDataThreshold {
		return v.hive.readBigData(v.DataOffset, actualSize)
	}

	return v.hive.readCellData(v.DataOffset, int(actualSize))
}

// String decodes the value data as a string for REG_SZ / REG_EXPAND_SZ types.
func (v *KeyValue) String() string {
	data, err := v.Data()
	if err != nil || data == nil {
		return ""
	}
	switch v.DataType {
	case RegSz, RegExpandSz, RegLink:
		return strings.TrimRight(decodeUTF16LE(data), "\x00")
	default:
		return ""
	}
}

// Uint32 decodes a REG_DWORD value. Returns (0, false) if not applicable.
func (v *KeyValue) Uint32() (uint32, bool) {
	data, err := v.Data()
	if err != nil || len(data) < 4 {
		return 0, false
	}
	return binary.LittleEndian.Uint32(data), true
}
