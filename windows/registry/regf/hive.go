package regf

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// Hive represents an opened offline Windows registry hive file. The zero value is not
// usable; create one with Open or OpenBytes.
type Hive struct {
	data      []byte
	baseBlock BaseBlock
	dirty     bool // set by mutating operations; gates base-block finalization in Bytes
}

// Open opens and parses a registry hive file from disk.
//
// Parameters:
//   - path (string): filesystem path to the hive file.
//
// Returns:
//   - A parsed Hive ready for queries.
//   - An error if the file cannot be read or has an invalid format.
func Open(path string) (*Hive, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("regf: open %s: %w", path, err)
	}
	return OpenBytes(data)
}

// OpenBytes parses a registry hive from an in-memory byte slice.
//
// Parameters:
//   - data ([]byte): raw hive file content.
//
// Returns:
//   - A parsed Hive ready for queries.
//   - An error if the data has an invalid format.
func OpenBytes(data []byte) (*Hive, error) {
	h := &Hive{data: data}

	if _, err := h.baseBlock.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("regf: %w", err)
	}

	return h, nil
}

// Close releases the hive data. The Hive is not usable after Close.
func (h *Hive) Close() error {
	h.data = nil
	return nil
}

// BaseBlock returns the parsed file header.
func (h *Hive) BaseBlock() *BaseBlock {
	return &h.baseBlock
}

// RootKey returns the root KeyNode of the hive.
//
// Returns:
//   - The root KeyNode.
//   - An error if the root cell cannot be read.
func (h *Hive) RootKey() (*KeyNode, error) {
	return h.readKeyNode(h.baseBlock.RootCellOffset)
}

// FindKey locates a registry key by path relative to the root key.
// Path components are separated by backslashes. A leading backslash is optional.
//
// Parameters:
//   - path (string): key path, e.g. "SAM\\Domains\\Account" or "ControlSet001\\Control\\Lsa\\JD".
//
// Returns:
//   - The KeyNode at the given path.
//   - An error if any component is not found.
func (h *Hive) FindKey(path string) (*KeyNode, error) {
	path = strings.TrimPrefix(path, "\\")
	root, err := h.RootKey()
	if err != nil {
		return nil, err
	}

	if path == "" {
		return root, nil
	}

	current := root
	for _, part := range strings.Split(path, "\\") {
		subkeys, err := current.SubKeys()
		if err != nil {
			return nil, fmt.Errorf("regf: enumerating subkeys of %q: %w", current.Name(), err)
		}
		found := false
		for _, sk := range subkeys {
			if strings.EqualFold(sk.Name(), part) {
				current = sk
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("regf: key %q not found under %q", part, current.Name())
		}
	}

	return current, nil
}

// EnumKey returns the names of all subkeys under the given path.
//
// Parameters:
//   - path (string): key path relative to the root key.
//
// Returns:
//   - A slice of subkey names.
//   - An error if the key is not found.
func (h *Hive) EnumKey(path string) ([]string, error) {
	nk, err := h.FindKey(path)
	if err != nil {
		return nil, err
	}
	subkeys, err := nk.SubKeys()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(subkeys))
	for i, sk := range subkeys {
		names[i] = sk.Name()
	}
	return names, nil
}

// EnumValues returns all value names under the given key path.
//
// Parameters:
//   - path (string): key path relative to the root key.
//
// Returns:
//   - A slice of value names (empty string for the default value).
//   - An error if the key is not found.
func (h *Hive) EnumValues(path string) ([]string, error) {
	nk, err := h.FindKey(path)
	if err != nil {
		return nil, err
	}
	values, err := nk.Values()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(values))
	for i, v := range values {
		names[i] = v.Name()
	}
	return names, nil
}

// GetValue reads a named value from the key at the given path.
//
// Parameters:
//   - keyPath (string): key path relative to the root key.
//   - valueName (string): value name; empty string for the default value.
//
// Returns:
//   - The data type (REG_* constant).
//   - The raw value data bytes.
//   - An error if the key or value is not found.
func (h *Hive) GetValue(keyPath, valueName string) (uint32, []byte, error) {
	nk, err := h.FindKey(keyPath)
	if err != nil {
		return 0, nil, err
	}
	v, err := nk.Value(valueName)
	if err != nil {
		return 0, nil, fmt.Errorf("regf: key %q: %w", keyPath, err)
	}
	data, err := v.Data()
	if err != nil {
		return 0, nil, fmt.Errorf("regf: key %q value %q: %w", keyPath, valueName, err)
	}
	return v.DataType, data, nil
}

// GetSecurity reads the self-relative SECURITY_DESCRIPTOR for the key at the given path.
//
// Parameters:
//   - keyPath (string): key path relative to the root key.
//
// Returns:
//   - The raw security descriptor bytes, or nil if the key has no SK record.
//   - An error if the key is not found or the SK record cannot be read.
func (h *Hive) GetSecurity(keyPath string) ([]byte, error) {
	nk, err := h.FindKey(keyPath)
	if err != nil {
		return nil, err
	}
	return nk.SecurityDescriptor()
}

// GetClass reads the class data for the key at the given path.
//
// Parameters:
//   - keyPath (string): key path relative to the root key.
//
// Returns:
//   - The raw class data bytes.
//   - An error if the key is not found or has no class data.
func (h *Hive) GetClass(keyPath string) ([]byte, error) {
	nk, err := h.FindKey(keyPath)
	if err != nil {
		return nil, err
	}
	data, err := nk.ClassData()
	if err != nil {
		return nil, fmt.Errorf("regf: key %q: %w", keyPath, err)
	}
	if data == nil {
		return nil, fmt.Errorf("regf: key %q has no class data", keyPath)
	}
	return data, nil
}

// readCellRaw reads a raw cell at the given offset, returning the cell data after
// the 4-byte size prefix. The offset is relative to the start of hive bins data.
func (h *Hive) readCellRaw(offset uint32) ([]byte, error) {
	absOffset := baseBlockSize + int(offset)
	if absOffset+4 > len(h.data) {
		return nil, fmt.Errorf("regf: cell offset 0x%X out of bounds (file size: %d)", offset, len(h.data))
	}

	cellSize := int32(binary.LittleEndian.Uint32(h.data[absOffset : absOffset+4]))
	if cellSize < 0 {
		cellSize = -cellSize
	}
	if cellSize < 4 {
		return nil, fmt.Errorf("regf: invalid cell size %d at offset 0x%X", cellSize, offset)
	}
	end := absOffset + int(cellSize)
	if end > len(h.data) {
		end = len(h.data)
	}

	return h.data[absOffset+4 : end], nil
}

// readCellData reads raw data from a cell, skipping the 4-byte size prefix and returning
// exactly 'size' bytes.
func (h *Hive) readCellData(offset uint32, size int) ([]byte, error) {
	absOffset := baseBlockSize + int(offset) + 4 // skip cell size prefix
	if absOffset+size > len(h.data) {
		return nil, fmt.Errorf("regf: data read at offset 0x%X+%d exceeds file bounds", offset, size)
	}
	result := make([]byte, size)
	copy(result, h.data[absOffset:absOffset+size])
	return result, nil
}

// readKeyNode reads and parses a KeyNode at the given cell offset.
func (h *Hive) readKeyNode(offset uint32) (*KeyNode, error) {
	cell, err := h.readCellRaw(offset)
	if err != nil {
		return nil, fmt.Errorf("regf: reading key node at 0x%X: %w", offset, err)
	}

	nk := NewKeyNode()
	if _, err := nk.Unmarshal(cell); err != nil {
		return nil, fmt.Errorf("regf: parsing key node at 0x%X: %w", offset, err)
	}
	nk.hive = h
	nk.offset = offset
	return nk, nil
}

// readValueList reads and parses a list of KeyValue records.
func (h *Hive) readValueList(offset, count uint32) ([]*KeyValue, error) {
	cell, err := h.readCellRaw(offset)
	if err != nil {
		return nil, fmt.Errorf("regf: reading value list at 0x%X: %w", offset, err)
	}

	values := make([]*KeyValue, 0, count)
	for i := uint32(0); i < count; i++ {
		off := int(i) * 4
		if off+4 > len(cell) {
			break
		}
		vkOffset := binary.LittleEndian.Uint32(cell[off : off+4])
		vk, err := h.readKeyValue(vkOffset)
		if err != nil {
			continue
		}
		values = append(values, vk)
	}
	return values, nil
}

// readSecurityDescriptor reads the SK record at the given cell offset and returns its
// embedded self-relative security descriptor bytes.
func (h *Hive) readSecurityDescriptor(offset uint32) ([]byte, error) {
	cell, err := h.readCellRaw(offset)
	if err != nil {
		return nil, fmt.Errorf("regf: reading security key at 0x%X: %w", offset, err)
	}
	sk := NewSecurityKey()
	if _, err := sk.Unmarshal(cell); err != nil {
		return nil, fmt.Errorf("regf: parsing security key at 0x%X: %w", offset, err)
	}
	return sk.SecurityDescriptor, nil
}

// readBigData reassembles a value's data from a big-data (db) record at the given offset.
// totalSize is the value's actual data size (ActualDataSize); the concatenated segments
// are truncated to it. If the cell is not a db record (e.g. a pre-1.4 large single cell),
// it falls back to reading totalSize bytes directly from the cell.
func (h *Hive) readBigData(offset, totalSize uint32) ([]byte, error) {
	cell, err := h.readCellRaw(offset)
	if err != nil {
		return nil, fmt.Errorf("regf: reading big data at 0x%X: %w", offset, err)
	}

	db := NewBigData()
	if _, err := db.Unmarshal(cell); err != nil {
		// Not a big-data record: treat the cell as a single large data cell.
		return h.readCellData(offset, int(totalSize))
	}

	listCell, err := h.readCellRaw(db.SegmentsListOffset)
	if err != nil {
		return nil, fmt.Errorf("regf: reading big-data segment list at 0x%X: %w", db.SegmentsListOffset, err)
	}

	result := make([]byte, 0, totalSize)
	for i := 0; i < int(db.NumberOfSegments); i++ {
		off := i * 4
		if off+4 > len(listCell) {
			break
		}
		segOffset := binary.LittleEndian.Uint32(listCell[off : off+4])
		segCell, err := h.readCellRaw(segOffset)
		if err != nil {
			return nil, fmt.Errorf("regf: reading big-data segment %d at 0x%X: %w", i, segOffset, err)
		}
		remaining := int(totalSize) - len(result)
		if remaining <= 0 {
			break
		}
		take := len(segCell)
		if take > bigDataThreshold {
			take = bigDataThreshold // each segment carries at most this many data bytes
		}
		if take > remaining {
			take = remaining
		}
		result = append(result, segCell[:take]...)
	}
	return result, nil
}

// readKeyValue reads and parses a KeyValue at the given cell offset.
func (h *Hive) readKeyValue(offset uint32) (*KeyValue, error) {
	cell, err := h.readCellRaw(offset)
	if err != nil {
		return nil, fmt.Errorf("regf: reading key value at 0x%X: %w", offset, err)
	}

	vk := NewKeyValue()
	if _, err := vk.Unmarshal(cell); err != nil {
		return nil, fmt.Errorf("regf: parsing key value at 0x%X: %w", offset, err)
	}
	vk.hive = h
	return vk, nil
}

// enumSubKeys collects all subkey KeyNode records from a subkeys list offset.
func (h *Hive) enumSubKeys(offset uint32) ([]*KeyNode, error) {
	offsets, err := h.collectKeyNodeOffsets(offset)
	if err != nil {
		return nil, err
	}

	subkeys := make([]*KeyNode, 0, len(offsets))
	for _, off := range offsets {
		nk, err := h.readKeyNode(off)
		if err != nil {
			continue
		}
		subkeys = append(subkeys, nk)
	}
	return subkeys, nil
}

// maxSubkeyListDepth bounds the recursion that resolves index-root (ri) subkey lists. Real
// hives nest ri at most one or two levels; this guards against a cyclic or maliciously
// deep ri list (which would otherwise recurse until the stack overflows).
const maxSubkeyListDepth = 32

// collectKeyNodeOffsets reads a subkey list and recursively resolves RI blocks.
func (h *Hive) collectKeyNodeOffsets(offset uint32) ([]uint32, error) {
	return h.collectKeyNodeOffsetsDepth(offset, 0)
}

func (h *Hive) collectKeyNodeOffsetsDepth(offset uint32, depth int) ([]uint32, error) {
	if depth > maxSubkeyListDepth {
		return nil, fmt.Errorf("regf: subkey list nesting too deep at 0x%X (possible cycle)", offset)
	}

	cell, err := h.readCellRaw(offset)
	if err != nil {
		return nil, fmt.Errorf("regf: reading subkey list at 0x%X: %w", offset, err)
	}

	list := NewSubKeyList()
	if _, err := list.Unmarshal(cell); err != nil {
		return nil, fmt.Errorf("regf: parsing subkey list at 0x%X: %w", offset, err)
	}

	if !list.IsIndexRoot() {
		return list.KeyNodeOffsets(), nil
	}

	// RI: each element points to another subkey list (LF/LH/LI). Recurse with a depth
	// bound so a cyclic ri cannot recurse without limit.
	var allOffsets []uint32
	for _, subListOffset := range list.KeyNodeOffsets() {
		subOffsets, err := h.collectKeyNodeOffsetsDepth(subListOffset, depth+1)
		if err != nil {
			continue
		}
		allOffsets = append(allOffsets, subOffsets...)
	}
	return allOffsets, nil
}
