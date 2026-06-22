package regf

import (
	"encoding/binary"
	"testing"
)

// buildCyclicRIHive builds a hive whose root's subkey list is an index-root (ri) cell with
// a single element that points to itself — a cycle that must not crash the parser.
func buildCyclicRIHive() []byte {
	a := &hiveAsm{}
	hb := &HiveBin{Signature: hbinSignature, Size: 4096}
	hdr, _ := hb.Marshal()
	a.bins = append(a.bins, hdr...)

	rootName := utf16le("ROOT")
	root := &KeyNode{
		Signature: nkSignature, Flags: KeyHiveEntry, NumberOfSubKeys: 1,
		ValuesListOffset: nullCellOffset, SecurityOffset: nullCellOffset, ClassNameOffset: nullCellOffset,
		Parent: nullCellOffset, KeyNameLength: uint16(len(rootName)), KeyNameRaw: rootName,
	}
	rootb, _ := root.Marshal()
	rootOff := a.addCell(rootb)
	rootContent := int(rootOff) + 4

	riOff := a.addCell(make([]byte, 8)) // sig(2) + count(2) + one 4-byte element
	binary.LittleEndian.PutUint16(a.bins[int(riOff)+4:], riSig)
	binary.LittleEndian.PutUint16(a.bins[int(riOff)+6:], 1)
	binary.LittleEndian.PutUint32(a.bins[int(riOff)+8:], riOff) // self-reference

	binary.LittleEndian.PutUint32(a.bins[rootContent+28:rootContent+32], riOff) // SubKeysListOffset

	if len(a.bins) < 4096 {
		a.bins = append(a.bins, make([]byte, 4096-len(a.bins))...)
	}
	bb := &BaseBlock{
		Signature: regfSignature, MajorVersion: 1, MinorVersion: 6, FileFormat: 1,
		RootCellOffset: rootOff, HiveBinsDataSize: uint32(len(a.bins)), ClusteringFactor: 1,
		PrimarySequenceNumber: 1, SecondarySequenceNumber: 1,
	}
	bbb, _ := bb.Marshal()
	return append(bbb, a.bins...)
}

// TestCyclicRIDoesNotRecurse verifies that enumerating a key with a self-referential ri
// subkey list terminates (bounded by maxSubkeyListDepth) instead of overflowing the stack.
func TestCyclicRIDoesNotRecurse(t *testing.T) {
	h, err := OpenBytes(buildCyclicRIHive())
	if err != nil {
		t.Fatalf("OpenBytes: %v", err)
	}
	// Must return without a stack overflow; the cyclic branch yields no keys.
	keys, err := h.EnumKey("")
	if err != nil {
		t.Fatalf("EnumKey: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("EnumKey on a cyclic ri = %v, want none", keys)
	}
	// DeleteKey walks the same structure on the write path; it must also terminate.
	if err := h.DeleteKey("", "anything"); err == nil {
		t.Error("DeleteKey of a missing key returned nil error")
	}
}
