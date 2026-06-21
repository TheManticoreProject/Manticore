package regf

import (
	"encoding/binary"
	"fmt"
	"strings"
	"unicode/utf16"
)

const (
	nkSignature     = 0x6B6E // "nk" in little-endian
	keyNodeMinSize  = 76
	nullCellOffset  = 0xFFFFFFFF
)

// Key node flags.
//
// Source: https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md#key-node
const (
	KeyVolatile   uint16 = 0x0001
	KeyHiveExit   uint16 = 0x0002
	KeyHiveEntry  uint16 = 0x0004
	KeyNoDelete   uint16 = 0x0008
	KeySymLink    uint16 = 0x0010
	KeyCompName   uint16 = 0x0020
	KeyPredefHndl uint16 = 0x0040
)

// KeyNode is a parsed NK (key node) record representing a registry key.
//
// Source: https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md#key-node
type KeyNode struct {
	// Signature (2 bytes): must be ASCII "nk" (0x6B6E little-endian).
	Signature uint16

	// Flags (2 bytes): bit mask of key flags.
	Flags uint16

	// LastWrittenTimestamp (8 bytes): FILETIME (UTC).
	LastWrittenTimestamp uint64

	// AccessBits (4 bytes): access tracking (Win10 RS1+); otherwise spare.
	AccessBits uint32

	// Parent (4 bytes): offset to parent key node.
	Parent uint32

	// NumberOfSubKeys (4 bytes): count of stable (non-volatile) subkeys.
	NumberOfSubKeys uint32

	// NumberOfVolatileSubKeys (4 bytes): count of volatile subkeys (no disk meaning).
	NumberOfVolatileSubKeys uint32

	// SubKeysListOffset (4 bytes): offset to subkey list (lf/lh/li/ri), or 0xFFFFFFFF.
	SubKeysListOffset uint32

	// VolatileSubKeysListOffset (4 bytes): no disk meaning.
	VolatileSubKeysListOffset uint32

	// NumberOfValues (4 bytes): count of values under this key.
	NumberOfValues uint32

	// ValuesListOffset (4 bytes): offset to values list, or 0xFFFFFFFF.
	ValuesListOffset uint32

	// SecurityOffset (4 bytes): offset to sk record.
	SecurityOffset uint32

	// ClassNameOffset (4 bytes): offset to cell containing class name data, or 0xFFFFFFFF.
	ClassNameOffset uint32

	// MaxSubKeyNameLength (4 bytes): largest subkey name length (with user/debug flags).
	MaxSubKeyNameLength uint32

	// MaxSubKeyClassNameLength (4 bytes): largest subkey class name length.
	MaxSubKeyClassNameLength uint32

	// MaxValueNameLength (4 bytes): largest value name length.
	MaxValueNameLength uint32

	// MaxValueDataSize (4 bytes): largest value data size.
	MaxValueDataSize uint32

	// WorkVar (4 bytes): cached subkey index (Win2000 only).
	WorkVar uint32

	// KeyNameLength (2 bytes): length of key name in bytes.
	KeyNameLength uint16

	// ClassNameLength (2 bytes): length of class name in bytes.
	ClassNameLength uint16

	// KeyNameRaw (variable): raw key name bytes.
	KeyNameRaw []byte

	// hive back-reference for navigation methods.
	hive   *Hive
	offset uint32
}

// NewKeyNode creates a new empty KeyNode.
func NewKeyNode() *KeyNode {
	return &KeyNode{}
}

// Unmarshal deserializes a KeyNode from cell data (after the 4-byte cell size prefix).
//
// Parameters:
//   - data ([]byte): cell data starting with the "nk" signature.
//
// Returns:
//   - The number of bytes consumed.
//   - An error if the data is too short or the signature is invalid.
func (k *KeyNode) Unmarshal(data []byte) (int, error) {
	if len(data) < keyNodeMinSize {
		return 0, fmt.Errorf("data too short for KeyNode: need %d bytes, got %d", keyNodeMinSize, len(data))
	}

	k.Signature = binary.LittleEndian.Uint16(data[0:2])
	if k.Signature != nkSignature {
		return 0, fmt.Errorf("invalid KeyNode Signature: 0x%04X (expected 0x%04X)", k.Signature, nkSignature)
	}

	k.Flags = binary.LittleEndian.Uint16(data[2:4])
	k.LastWrittenTimestamp = binary.LittleEndian.Uint64(data[4:12])
	k.AccessBits = binary.LittleEndian.Uint32(data[12:16])
	k.Parent = binary.LittleEndian.Uint32(data[16:20])
	k.NumberOfSubKeys = binary.LittleEndian.Uint32(data[20:24])
	k.NumberOfVolatileSubKeys = binary.LittleEndian.Uint32(data[24:28])
	k.SubKeysListOffset = binary.LittleEndian.Uint32(data[28:32])
	k.VolatileSubKeysListOffset = binary.LittleEndian.Uint32(data[32:36])
	k.NumberOfValues = binary.LittleEndian.Uint32(data[36:40])
	k.ValuesListOffset = binary.LittleEndian.Uint32(data[40:44])
	k.SecurityOffset = binary.LittleEndian.Uint32(data[44:48])
	k.ClassNameOffset = binary.LittleEndian.Uint32(data[48:52])
	k.MaxSubKeyNameLength = binary.LittleEndian.Uint32(data[52:56])
	k.MaxSubKeyClassNameLength = binary.LittleEndian.Uint32(data[56:60])
	k.MaxValueNameLength = binary.LittleEndian.Uint32(data[60:64])
	k.MaxValueDataSize = binary.LittleEndian.Uint32(data[64:68])
	k.WorkVar = binary.LittleEndian.Uint32(data[68:72])
	k.KeyNameLength = binary.LittleEndian.Uint16(data[72:74])
	k.ClassNameLength = binary.LittleEndian.Uint16(data[74:76])

	nameEnd := keyNodeMinSize + int(k.KeyNameLength)
	if nameEnd > len(data) {
		nameEnd = len(data)
	}
	k.KeyNameRaw = make([]byte, nameEnd-keyNodeMinSize)
	copy(k.KeyNameRaw, data[keyNodeMinSize:nameEnd])

	return nameEnd, nil
}

// Marshal serializes the KeyNode to binary data.
//
// Returns:
//   - A byte slice containing the serialized KeyNode.
//   - An error if serialization fails.
func (k *KeyNode) Marshal() ([]byte, error) {
	size := keyNodeMinSize + len(k.KeyNameRaw)
	buf := make([]byte, size)

	binary.LittleEndian.PutUint16(buf[0:2], k.Signature)
	binary.LittleEndian.PutUint16(buf[2:4], k.Flags)
	binary.LittleEndian.PutUint64(buf[4:12], k.LastWrittenTimestamp)
	binary.LittleEndian.PutUint32(buf[12:16], k.AccessBits)
	binary.LittleEndian.PutUint32(buf[16:20], k.Parent)
	binary.LittleEndian.PutUint32(buf[20:24], k.NumberOfSubKeys)
	binary.LittleEndian.PutUint32(buf[24:28], k.NumberOfVolatileSubKeys)
	binary.LittleEndian.PutUint32(buf[28:32], k.SubKeysListOffset)
	binary.LittleEndian.PutUint32(buf[32:36], k.VolatileSubKeysListOffset)
	binary.LittleEndian.PutUint32(buf[36:40], k.NumberOfValues)
	binary.LittleEndian.PutUint32(buf[40:44], k.ValuesListOffset)
	binary.LittleEndian.PutUint32(buf[44:48], k.SecurityOffset)
	binary.LittleEndian.PutUint32(buf[48:52], k.ClassNameOffset)
	binary.LittleEndian.PutUint32(buf[52:56], k.MaxSubKeyNameLength)
	binary.LittleEndian.PutUint32(buf[56:60], k.MaxSubKeyClassNameLength)
	binary.LittleEndian.PutUint32(buf[60:64], k.MaxValueNameLength)
	binary.LittleEndian.PutUint32(buf[64:68], k.MaxValueDataSize)
	binary.LittleEndian.PutUint32(buf[68:72], k.WorkVar)
	binary.LittleEndian.PutUint16(buf[72:74], k.KeyNameLength)
	binary.LittleEndian.PutUint16(buf[74:76], k.ClassNameLength)
	copy(buf[keyNodeMinSize:], k.KeyNameRaw)

	return buf, nil
}

// Name returns the decoded key name as a Go string.
func (k *KeyNode) Name() string {
	if k.Flags&KeyCompName != 0 {
		return string(k.KeyNameRaw)
	}
	return decodeUTF16LE(k.KeyNameRaw)
}

// IsRoot reports whether this key is a root key (KEY_HIVE_ENTRY flag set).
func (k *KeyNode) IsRoot() bool {
	return k.Flags&KeyHiveEntry != 0
}

// SubKeys returns all subkey KeyNodes under this key.
func (k *KeyNode) SubKeys() ([]*KeyNode, error) {
	if k.hive == nil {
		return nil, fmt.Errorf("KeyNode not attached to a hive")
	}
	if k.NumberOfSubKeys == 0 || k.SubKeysListOffset == nullCellOffset {
		return nil, nil
	}
	return k.hive.enumSubKeys(k.SubKeysListOffset)
}

// Values returns all KeyValue records under this key.
func (k *KeyNode) Values() ([]*KeyValue, error) {
	if k.hive == nil {
		return nil, fmt.Errorf("KeyNode not attached to a hive")
	}
	if k.NumberOfValues == 0 || k.ValuesListOffset == nullCellOffset {
		return nil, nil
	}
	return k.hive.readValueList(k.ValuesListOffset, k.NumberOfValues)
}

// Value returns the named value under this key, or an error if not found.
func (k *KeyNode) Value(name string) (*KeyValue, error) {
	values, err := k.Values()
	if err != nil {
		return nil, err
	}
	for _, v := range values {
		if strings.EqualFold(v.Name(), name) {
			return v, nil
		}
		if name == "" && v.NameLength == 0 {
			return v, nil
		}
	}
	return nil, fmt.Errorf("value %q not found", name)
}

// ClassData returns the class data bytes for this key, or nil if not set.
func (k *KeyNode) ClassData() ([]byte, error) {
	if k.hive == nil {
		return nil, fmt.Errorf("KeyNode not attached to a hive")
	}
	if k.ClassNameOffset == nullCellOffset || k.ClassNameLength == 0 {
		return nil, nil
	}
	return k.hive.readCellData(k.ClassNameOffset, int(k.ClassNameLength))
}

// decodeUTF16LE decodes UTF-16LE bytes to a Go string, dropping trailing NULs.
func decodeUTF16LE(data []byte) string {
	if len(data) < 2 {
		return string(data)
	}
	units := make([]uint16, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		units = append(units, binary.LittleEndian.Uint16(data[i:i+2]))
	}
	for len(units) > 0 && units[len(units)-1] == 0 {
		units = units[:len(units)-1]
	}
	return string(utf16.Decode(units))
}
