package regf

import (
	"encoding/binary"
	"fmt"
)

// walkCells iterates every cell in the hive, in order, across all hive bins. For each cell
// it calls visit with the cell's offset (relative to hive-bins data), whether the cell is
// allocated (size prefix negative) or free (size prefix positive), and the cell content
// after the 4-byte size prefix. Iteration stops early if visit returns false.
//
// The walk is defensive against malformed input: it stops a bin (or the whole walk) on an
// invalid HBIN header, a zero or sub-minimal cell size, or any size that would run past
// the bin or the file, rather than reading out of bounds or looping forever.
func (h *Hive) walkCells(visit func(offset uint32, allocated bool, content []byte) bool) {
	total := int(h.baseBlock.HiveBinsDataSize)
	binStart := 0
	for binStart < total {
		hbinAbs := baseBlockSize + binStart
		if hbinAbs+hbinHeaderSize > len(h.data) {
			return
		}
		hb := NewHiveBin()
		if _, err := hb.Unmarshal(h.data[hbinAbs:]); err != nil {
			return // not a hive bin: stop walking
		}
		binSize := int(hb.Size)
		if binSize < hbinHeaderSize {
			return
		}
		binEnd := binStart + binSize
		if binEnd > total {
			binEnd = total
		}

		cellPos := binStart + hbinHeaderSize
		for cellPos+4 <= binEnd {
			sizeAbs := baseBlockSize + cellPos
			if sizeAbs+4 > len(h.data) {
				break
			}
			raw := int32(binary.LittleEndian.Uint32(h.data[sizeAbs : sizeAbs+4]))
			if raw == 0 {
				break // free-space terminator / zero padding
			}
			allocated := raw < 0
			cellSize := int(raw)
			if allocated {
				cellSize = int(-raw)
			}
			if cellSize < 4 || cellPos+cellSize > binEnd {
				break
			}
			contentEnd := sizeAbs + cellSize
			if contentEnd > len(h.data) {
				contentEnd = len(h.data)
			}
			if !visit(uint32(cellPos), allocated, h.data[sizeAbs+4:contentEnd]) {
				return
			}
			cellPos += cellSize
		}
		binStart += binSize
	}
}

// RecoverDeletedKeys scans the hive's unallocated (free) cells for key-node (NK) records
// that survived deletion and returns them. A deleted key is one whose cell was freed (its
// size prefix flipped positive) but whose bytes have not yet been overwritten.
//
// The returned KeyNodes are attached to the hive so Name and the other field accessors
// work, but navigation from them (SubKeys, Values) is unreliable: a deleted key's child
// list and value offsets may point at cells that have since been reallocated. Treat the
// result as recovered metadata, not as live tree nodes.
func (h *Hive) RecoverDeletedKeys() ([]*KeyNode, error) {
	if h.data == nil {
		return nil, fmt.Errorf("regf: hive is closed")
	}
	var out []*KeyNode
	h.walkCells(func(offset uint32, allocated bool, content []byte) bool {
		if allocated || len(content) < 2 {
			return true
		}
		if binary.LittleEndian.Uint16(content[0:2]) != nkSignature {
			return true
		}
		nk := NewKeyNode()
		if _, err := nk.Unmarshal(content); err != nil {
			return true // signature matched but the record is truncated/garbled: skip
		}
		nk.hive = h
		nk.offset = offset
		out = append(out, nk)
		return true
	})
	return out, nil
}

// RecoverDeletedValues scans the hive's unallocated (free) cells for value (VK) records
// that survived deletion and returns them. As with RecoverDeletedKeys, treat these as
// recovered metadata; a recovered value's external data cell may have been reallocated, so
// Data may return stale or unrelated bytes (inline values are self-contained and safe).
func (h *Hive) RecoverDeletedValues() ([]*KeyValue, error) {
	if h.data == nil {
		return nil, fmt.Errorf("regf: hive is closed")
	}
	var out []*KeyValue
	h.walkCells(func(offset uint32, allocated bool, content []byte) bool {
		if allocated || len(content) < 2 {
			return true
		}
		if binary.LittleEndian.Uint16(content[0:2]) != vkSignature {
			return true
		}
		vk := NewKeyValue()
		if _, err := vk.Unmarshal(content); err != nil {
			return true
		}
		vk.hive = h
		out = append(out, vk)
		return true
	})
	return out, nil
}
