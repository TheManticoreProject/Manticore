package regf

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// This file implements write support for hives: a cell allocator over the hive's mutable
// buffer, plus value create/update/delete. Mutations are applied directly to the in-memory
// image (h.data); navigation always re-reads from that buffer, so changes are visible
// immediately. Bytes/Save finalize the base block (sequence numbers and checksum).
//
// Scope (phase 1): values on existing keys. Key creation/deletion (subkey-list editing)
// is separate. Value names are written as compressed (Latin-1) strings.

// align8 rounds n up to a multiple of 8 (the hive cell alignment).
func align8(n int) int {
	if r := n % 8; r != 0 {
		return n + 8 - r
	}
	return n
}

// align4096 rounds n up to a multiple of 4096 (the hive bin granularity).
func align4096(n int) int {
	if r := n % 4096; r != 0 {
		return n + 4096 - r
	}
	return n
}

// cellSizeAt returns the signed size prefix of the cell at the given hive-bins-data offset.
func (h *Hive) cellSizeAt(offset uint32) (int32, bool) {
	abs := baseBlockSize + int(offset)
	if abs+4 > len(h.data) {
		return 0, false
	}
	return int32(binary.LittleEndian.Uint32(h.data[abs : abs+4])), true
}

// setCellSize writes the signed size prefix of the cell at the given offset (negative =
// allocated, positive = free).
func (h *Hive) setCellSize(offset uint32, size int32) {
	abs := baseBlockSize + int(offset)
	binary.LittleEndian.PutUint32(h.data[abs:abs+4], uint32(size))
}

// writeCellContent copies data into the content area of the cell at offset (after the
// 4-byte size prefix). The caller must ensure the cell is large enough.
func (h *Hive) writeCellContent(offset uint32, data []byte) {
	abs := baseBlockSize + int(offset) + 4
	copy(h.data[abs:abs+len(data)], data)
}

// allocateCell returns the offset of an allocated cell whose content area is at least
// contentSize bytes. It first-fits an existing free cell (splitting off the remainder when
// large enough), otherwise appends a new hive bin. The cell's content is zeroed.
func (h *Hive) allocateCell(contentSize int) (uint32, error) {
	if h.data == nil {
		return 0, fmt.Errorf("regf: hive is closed")
	}
	need := align8(4 + contentSize)

	found := uint32(nullCellOffset)
	foundSize := 0
	h.walkCells(func(off uint32, allocated bool, content []byte) bool {
		if allocated {
			return true
		}
		total := len(content) + 4
		if total >= need {
			found, foundSize = off, total
			return false // first fit
		}
		return true
	})
	if found != nullCellOffset {
		h.claimFreeCell(found, foundSize, need)
		return found, nil
	}
	return h.growForCell(need)
}

// claimFreeCell marks a free cell allocated, splitting off the remainder as a new free cell
// when it is large enough to stand alone, and zeroes the allocated content.
func (h *Hive) claimFreeCell(offset uint32, total, need int) {
	remainder := total - need
	allocSize := need
	if remainder < 8 {
		allocSize = total // too small to split: absorb the remainder
		remainder = 0
	}
	h.setCellSize(offset, -int32(allocSize))
	// zero the content area
	abs := baseBlockSize + int(offset) + 4
	for i := abs; i < abs+allocSize-4; i++ {
		h.data[i] = 0
	}
	if remainder >= 8 {
		freeOff := offset + uint32(allocSize)
		h.setCellSize(freeOff, int32(remainder))
	}
}

// growForCell appends a new hive bin large enough to hold a cell with content area `need`
// (need is the full cell size, size-prefix included), allocates that cell at the start of
// the bin, and returns its offset.
func (h *Hive) growForCell(need int) (uint32, error) {
	binSize := align4096(hbinHeaderSize + need)
	binOffset := h.baseBlock.HiveBinsDataSize

	bin := make([]byte, binSize)
	hb := &HiveBin{Signature: hbinSignature, Offset: binOffset, Size: uint32(binSize)}
	hdr, _ := hb.Marshal()
	copy(bin, hdr)

	// Allocated cell at the start of the bin's cell area.
	binary.LittleEndian.PutUint32(bin[hbinHeaderSize:hbinHeaderSize+4], uint32(-int32(need)))
	cellOffset := binOffset + hbinHeaderSize
	// Remainder (if any) as a single free cell.
	if rem := binSize - hbinHeaderSize - need; rem >= 8 {
		binary.LittleEndian.PutUint32(bin[hbinHeaderSize+need:hbinHeaderSize+need+4], uint32(int32(rem)))
	} else if rem > 0 {
		// Fold the small remainder into the allocated cell.
		binary.LittleEndian.PutUint32(bin[hbinHeaderSize:hbinHeaderSize+4], uint32(-int32(need+rem)))
	}

	h.data = append(h.data, bin...)
	h.baseBlock.HiveBinsDataSize += uint32(binSize)
	binary.LittleEndian.PutUint32(h.data[40:44], h.baseBlock.HiveBinsDataSize)
	return cellOffset, nil
}

// freeCell marks the cell at offset as unallocated (positive size prefix). Adjacent free
// cells are not coalesced.
func (h *Hive) freeCell(offset uint32) {
	if sz, ok := h.cellSizeAt(offset); ok && sz < 0 {
		h.setCellSize(offset, -sz)
	}
}

// valueListOffsets reads the VK cell offsets from a key's value list.
func (h *Hive) valueListOffsets(listOffset, count uint32) []uint32 {
	if listOffset == nullCellOffset || count == 0 {
		return nil
	}
	abs := baseBlockSize + int(listOffset) + 4
	offsets := make([]uint32, 0, count)
	for i := uint32(0); i < count; i++ {
		p := abs + int(i)*4
		if p+4 > len(h.data) {
			break
		}
		offsets = append(offsets, binary.LittleEndian.Uint32(h.data[p:p+4]))
	}
	return offsets
}

// buildValueCell builds and writes a VK record (and an external data cell when needed) for
// the given value, returning the VK cell offset.
func (h *Hive) buildValueCell(name string, dataType uint32, data []byte) (uint32, error) {
	vk := &KeyValue{
		Signature:  vkSignature,
		NameLength: uint16(len(name)),
		DataType:   dataType,
		Flags:      ValueCompName,
		NameRaw:    []byte(name),
	}
	if len(data) <= 4 {
		var buf [4]byte
		copy(buf[:], data)
		vk.DataSize = 0x80000000 | uint32(len(data))
		vk.DataOffset = binary.LittleEndian.Uint32(buf[:])
	} else {
		dataOff, err := h.allocateCell(len(data))
		if err != nil {
			return 0, err
		}
		h.writeCellContent(dataOff, data)
		vk.DataSize = uint32(len(data))
		vk.DataOffset = dataOff
	}
	vkBytes, _ := vk.Marshal()
	vkOff, err := h.allocateCell(len(vkBytes))
	if err != nil {
		return 0, err
	}
	h.writeCellContent(vkOff, vkBytes)
	return vkOff, nil
}

// freeValueCell frees a VK cell and its external data cell (if any).
func (h *Hive) freeValueCell(vkOffset uint32) {
	if vk, err := h.readKeyValue(vkOffset); err == nil && !vk.IsInline() && vk.ActualDataSize() > 0 {
		h.freeCell(vk.DataOffset)
	}
	h.freeCell(vkOffset)
}

// SetValue creates or replaces a value on the key at keyPath. data is stored inline when it
// is 4 bytes or smaller, otherwise in its own data cell. The value name is stored as a
// compressed (Latin-1) string. The change is applied to the in-memory image; call Bytes or
// Save to obtain the finalized hive.
func (h *Hive) SetValue(keyPath, name string, dataType uint32, data []byte) error {
	if h.data == nil {
		return fmt.Errorf("regf: hive is closed")
	}
	nk, err := h.FindKey(keyPath)
	if err != nil {
		return err
	}
	nkContent := baseBlockSize + int(nk.offset) + 4

	offsets := h.valueListOffsets(nk.ValuesListOffset, nk.NumberOfValues)
	existing := -1
	for i, off := range offsets {
		if vk, err := h.readKeyValue(off); err == nil && strings.EqualFold(vk.Name(), name) {
			existing = i
			break
		}
	}

	vkOff, err := h.buildValueCell(name, dataType, data)
	if err != nil {
		return err
	}

	if existing >= 0 {
		// Replace the slot in place; free the old VK cell (+ its data cell).
		oldVK := offsets[existing]
		listContent := baseBlockSize + int(nk.ValuesListOffset) + 4
		binary.LittleEndian.PutUint32(h.data[listContent+existing*4:], vkOff)
		h.freeValueCell(oldVK)
		return nil
	}

	// Append: grow the value list by one slot.
	newCount := nk.NumberOfValues + 1
	newList, err := h.allocateCell(int(newCount) * 4)
	if err != nil {
		return err
	}
	lc := baseBlockSize + int(newList) + 4
	for i, off := range offsets {
		binary.LittleEndian.PutUint32(h.data[lc+i*4:], off)
	}
	binary.LittleEndian.PutUint32(h.data[lc+len(offsets)*4:], vkOff)
	if nk.ValuesListOffset != nullCellOffset {
		h.freeCell(nk.ValuesListOffset)
	}
	binary.LittleEndian.PutUint32(h.data[nkContent+36:nkContent+40], newCount) // NumberOfValues
	binary.LittleEndian.PutUint32(h.data[nkContent+40:nkContent+44], newList)  // ValuesListOffset
	return nil
}

// DeleteValue removes a named value from the key at keyPath, freeing its cells. It returns
// an error if the key or value does not exist.
func (h *Hive) DeleteValue(keyPath, name string) error {
	if h.data == nil {
		return fmt.Errorf("regf: hive is closed")
	}
	nk, err := h.FindKey(keyPath)
	if err != nil {
		return err
	}
	nkContent := baseBlockSize + int(nk.offset) + 4

	offsets := h.valueListOffsets(nk.ValuesListOffset, nk.NumberOfValues)
	idx := -1
	for i, off := range offsets {
		if vk, err := h.readKeyValue(off); err == nil && strings.EqualFold(vk.Name(), name) {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("regf: value %q not found under %q", name, keyPath)
	}

	h.freeValueCell(offsets[idx])
	newCount := nk.NumberOfValues - 1
	if newCount == 0 {
		h.freeCell(nk.ValuesListOffset)
		binary.LittleEndian.PutUint32(h.data[nkContent+36:nkContent+40], 0)
		binary.LittleEndian.PutUint32(h.data[nkContent+40:nkContent+44], nullCellOffset)
		return nil
	}

	newList, err := h.allocateCell(int(newCount) * 4)
	if err != nil {
		return err
	}
	lc := baseBlockSize + int(newList) + 4
	j := 0
	for i, off := range offsets {
		if i == idx {
			continue
		}
		binary.LittleEndian.PutUint32(h.data[lc+j*4:], off)
		j++
	}
	h.freeCell(nk.ValuesListOffset)
	binary.LittleEndian.PutUint32(h.data[nkContent+36:nkContent+40], newCount)
	binary.LittleEndian.PutUint32(h.data[nkContent+40:nkContent+44], newList)
	return nil
}

// finalize updates the base block to reflect mutations: it bumps the (equal) sequence
// numbers and recomputes the XOR-32 header checksum.
func (h *Hive) finalize() {
	seq := h.baseBlock.PrimarySequenceNumber + 1
	binary.LittleEndian.PutUint32(h.data[4:8], seq)
	binary.LittleEndian.PutUint32(h.data[8:12], seq)
	h.baseBlock.PrimarySequenceNumber = seq
	h.baseBlock.SecondarySequenceNumber = seq

	var cs uint32
	for i := 0; i < 508; i += 4 {
		cs ^= binary.LittleEndian.Uint32(h.data[i : i+4])
	}
	switch cs {
	case 0:
		cs = 1
	case 0xFFFFFFFF:
		cs = 0xFFFFFFFE
	}
	binary.LittleEndian.PutUint32(h.data[508:512], cs)
}

// Bytes finalizes the hive (sequence numbers and checksum) and returns a copy of the
// current image, suitable for writing to disk or re-opening with OpenBytes.
func (h *Hive) Bytes() ([]byte, error) {
	if h.data == nil {
		return nil, fmt.Errorf("regf: hive is closed")
	}
	h.finalize()
	return append([]byte(nil), h.data...), nil
}

// Save finalizes the hive and writes it to path.
func (h *Hive) Save(path string) error {
	b, err := h.Bytes()
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
