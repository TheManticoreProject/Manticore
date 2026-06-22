package regf

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

// This file extends write support (phase 2) with key creation and deletion, which edit a
// parent's subkey list. Leaf list types (lf, lh, li) are supported; an index-root (ri)
// subkey list — only produced for keys with very many subkeys — is rejected for edits.
// Subkey lists are kept sorted case-insensitively, as Windows expects, and lh name hashes
// are recomputed. Key names are written compressed (Latin-1).

// subkeyEntry is one subkey-list element: the key-node offset and the subkey name (used to
// recompute the list element's hint/hash and to keep the list sorted).
type subkeyEntry struct {
	offset uint32
	name   string
}

// lhHash computes the "lh" subkey-list name hash: fold the uppercased name with a factor
// of 37. (Matches the on-disk hash for ASCII names.)
func lhHash(name string) uint32 {
	var h uint32
	for _, r := range strings.ToUpper(name) {
		h = h*37 + uint32(r)
	}
	return h
}

// buildSubkeyList serializes a subkey list of the given leaf signature (lf/lh/li) from
// entries, recomputing each element's hint (lf) or hash (lh).
func buildSubkeyList(signature uint16, entries []subkeyEntry) []byte {
	elemSize := 4
	if signature == lfSig || signature == lhSig {
		elemSize = 8
	}
	buf := make([]byte, 4+len(entries)*elemSize)
	binary.LittleEndian.PutUint16(buf[0:2], signature)
	binary.LittleEndian.PutUint16(buf[2:4], uint16(len(entries)))
	for i, e := range entries {
		p := 4 + i*elemSize
		binary.LittleEndian.PutUint32(buf[p:p+4], e.offset)
		switch signature {
		case lfSig:
			for j := 0; j < 4 && j < len(e.name); j++ {
				buf[p+4+j] = e.name[j] // 4-char name hint
			}
		case lhSig:
			binary.LittleEndian.PutUint32(buf[p+4:p+8], lhHash(e.name))
		}
	}
	return buf
}

// bumpSKRefCount adjusts an SK record's reference count by delta, freeing the SK cell if
// the count reaches zero on a decrement. It is a no-op for a null/invalid security offset.
func (h *Hive) bumpSKRefCount(skOffset uint32, delta int) {
	if skOffset == nullCellOffset {
		return
	}
	c := baseBlockSize + int(skOffset) + 4
	if c+16 > len(h.data) || binary.LittleEndian.Uint16(h.data[c:c+2]) != skSignature {
		return
	}
	rc := int(binary.LittleEndian.Uint32(h.data[c+12:c+16])) + delta
	if rc < 0 {
		rc = 0
	}
	binary.LittleEndian.PutUint32(h.data[c+12:c+16], uint32(rc))
	if rc == 0 && delta < 0 {
		h.freeCell(skOffset)
	}
}

// CreateKey creates an empty subkey `name` under the key at parentPath. The new key
// inherits the parent's security descriptor (sharing its SK record). It errors if the
// parent does not exist, the subkey already exists, or the parent's subkey list is an
// index root (ri). The change is applied in memory; call Bytes or Save to finalize.
func (h *Hive) CreateKey(parentPath, name string) error {
	if h.data == nil {
		return fmt.Errorf("regf: hive is closed")
	}
	parent, err := h.FindKey(parentPath)
	if err != nil {
		return err
	}
	if subs, _ := parent.SubKeys(); subs != nil {
		for _, sk := range subs {
			if strings.EqualFold(sk.Name(), name) {
				return fmt.Errorf("regf: key %q already exists under %q", name, parentPath)
			}
		}
	}

	nk := &KeyNode{
		Signature:         nkSignature,
		Flags:             KeyCompName,
		SubKeysListOffset: nullCellOffset,
		ValuesListOffset:  nullCellOffset,
		ClassNameOffset:   nullCellOffset,
		SecurityOffset:    parent.SecurityOffset,
		Parent:            parent.offset,
		KeyNameLength:     uint16(len(name)),
		KeyNameRaw:        []byte(name),
	}
	nkBytes, _ := nk.Marshal()
	nkOff, err := h.allocateCell(len(nkBytes))
	if err != nil {
		return err
	}
	h.writeCellContent(nkOff, nkBytes)
	h.bumpSKRefCount(parent.SecurityOffset, +1)

	if err := h.addSubkeyToList(parent, nkOff, name); err != nil {
		return err
	}
	// Stamp both keys and keep the parent's "largest subkey name" hint covering the new key.
	h.touchKey(parent.offset)
	h.touchKey(nkOff)
	pc := baseBlockSize + int(parent.offset) + 4
	h.growU32(pc+52, uint32(len(name))) // MaxSubKeyNameLength
	h.dirty = true
	return nil
}

// addSubkeyToList inserts a new subkey (offset+name) into the parent's subkey list, keeping
// it sorted, and updates the parent's NumberOfSubKeys and SubKeysListOffset.
func (h *Hive) addSubkeyToList(parent *KeyNode, newOffset uint32, newName string) error {
	pc := baseBlockSize + int(parent.offset) + 4

	if parent.SubKeysListOffset == nullCellOffset || parent.NumberOfSubKeys == 0 {
		list := buildSubkeyList(lhSig, []subkeyEntry{{newOffset, newName}})
		lo, err := h.allocateCell(len(list))
		if err != nil {
			return err
		}
		h.writeCellContent(lo, list)
		binary.LittleEndian.PutUint32(h.data[pc+20:pc+24], 1)
		binary.LittleEndian.PutUint32(h.data[pc+28:pc+32], lo)
		return nil
	}

	entries, signature, err := h.readSubkeyEntries(parent.SubKeysListOffset)
	if err != nil {
		return err
	}
	entries = append(entries, subkeyEntry{newOffset, newName})
	sortSubkeyEntries(entries)

	list := buildSubkeyList(signature, entries)
	lo, err := h.allocateCell(len(list))
	if err != nil {
		return err
	}
	h.writeCellContent(lo, list)
	h.freeCell(parent.SubKeysListOffset)
	binary.LittleEndian.PutUint32(h.data[pc+20:pc+24], parent.NumberOfSubKeys+1)
	binary.LittleEndian.PutUint32(h.data[pc+28:pc+32], lo)
	return nil
}

// readSubkeyEntries reads a parent's subkey list into entries (offset+name) and returns the
// list's leaf signature. It errors on an index-root (ri) list.
func (h *Hive) readSubkeyEntries(listOffset uint32) ([]subkeyEntry, uint16, error) {
	cell, err := h.readCellRaw(listOffset)
	if err != nil {
		return nil, 0, err
	}
	list := NewSubKeyList()
	if _, err := list.Unmarshal(cell); err != nil {
		return nil, 0, err
	}
	if list.IsIndexRoot() {
		return nil, 0, fmt.Errorf("regf: editing an index-root (ri) subkey list is not supported")
	}
	var entries []subkeyEntry
	for _, off := range list.KeyNodeOffsets() {
		// Preserve every referenced subkey, even one whose node cannot be read: dropping it
		// here would silently delete that subkey from the parent when the list is rebuilt.
		// An unreadable node keeps an empty name (its hint/hash is then approximate).
		name := ""
		if sub, err := h.readKeyNode(off); err == nil {
			name = sub.Name()
		}
		entries = append(entries, subkeyEntry{off, name})
	}
	return entries, list.Signature, nil
}

func sortSubkeyEntries(entries []subkeyEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToUpper(entries[i].name) < strings.ToUpper(entries[j].name)
	})
}

// DeleteKey removes the subkey `name` under parentPath and everything beneath it
// (recursively freeing descendant keys, values, and their cells). It errors if the key is
// not found or the parent's subkey list is an index root (ri).
func (h *Hive) DeleteKey(parentPath, name string) error {
	if h.data == nil {
		return fmt.Errorf("regf: hive is closed")
	}
	parent, err := h.FindKey(parentPath)
	if err != nil {
		return err
	}
	if parent.SubKeysListOffset == nullCellOffset || parent.NumberOfSubKeys == 0 {
		return fmt.Errorf("regf: key %q not found under %q", name, parentPath)
	}

	entries, signature, err := h.readSubkeyEntries(parent.SubKeysListOffset)
	if err != nil {
		return err
	}
	targetOffset := uint32(nullCellOffset)
	remaining := entries[:0:0]
	for _, e := range entries {
		if targetOffset == nullCellOffset && strings.EqualFold(e.name, name) {
			targetOffset = e.offset
			continue
		}
		remaining = append(remaining, e)
	}
	if targetOffset == nullCellOffset {
		return fmt.Errorf("regf: key %q not found under %q", name, parentPath)
	}

	h.freeKeyRecursive(targetOffset)

	pc := baseBlockSize + int(parent.offset) + 4
	if len(remaining) == 0 {
		h.freeCell(parent.SubKeysListOffset)
		binary.LittleEndian.PutUint32(h.data[pc+20:pc+24], 0)
		binary.LittleEndian.PutUint32(h.data[pc+28:pc+32], nullCellOffset)
		h.touchKey(parent.offset)
		h.dirty = true
		return nil
	}
	list := buildSubkeyList(signature, remaining)
	lo, err := h.allocateCell(len(list))
	if err != nil {
		return err
	}
	h.writeCellContent(lo, list)
	h.freeCell(parent.SubKeysListOffset)
	binary.LittleEndian.PutUint32(h.data[pc+20:pc+24], uint32(len(remaining)))
	binary.LittleEndian.PutUint32(h.data[pc+28:pc+32], lo)
	h.touchKey(parent.offset)
	h.dirty = true
	return nil
}

// maxKeyTreeDepth bounds the recursion when freeing a key subtree, guarding against a
// cyclic subkey graph in a corrupt hive (Windows limits key depth well below this).
const maxKeyTreeDepth = 512

// freeKeyRecursive frees a key node and everything it owns: its values (and their data
// cells), its value list, its subkeys (recursively) and subkey list, its class cell, and
// decrements its shared SK record's reference count.
func (h *Hive) freeKeyRecursive(nkOffset uint32) {
	h.freeKeyRecursiveDepth(nkOffset, 0)
}

func (h *Hive) freeKeyRecursiveDepth(nkOffset uint32, depth int) {
	if depth > maxKeyTreeDepth {
		h.freeCell(nkOffset)
		return
	}
	nk, err := h.readKeyNode(nkOffset)
	if err != nil {
		h.freeCell(nkOffset)
		return
	}
	if nk.ValuesListOffset != nullCellOffset && nk.NumberOfValues > 0 {
		for _, vo := range h.valueListOffsets(nk.ValuesListOffset, nk.NumberOfValues) {
			h.freeValueCell(vo)
		}
		h.freeCell(nk.ValuesListOffset)
	}
	if nk.SubKeysListOffset != nullCellOffset && nk.NumberOfSubKeys > 0 {
		if offs, err := h.collectKeyNodeOffsets(nk.SubKeysListOffset); err == nil {
			for _, so := range offs {
				h.freeKeyRecursiveDepth(so, depth+1)
			}
		}
		h.freeCell(nk.SubKeysListOffset)
	}
	if nk.ClassNameOffset != nullCellOffset {
		h.freeCell(nk.ClassNameOffset)
	}
	h.bumpSKRefCount(nk.SecurityOffset, -1)
	h.freeCell(nkOffset)
}
