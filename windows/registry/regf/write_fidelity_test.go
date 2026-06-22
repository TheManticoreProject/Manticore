package regf

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// TestBytesNoOpWhenUnmodified verifies that opening a hive and calling Bytes without any
// mutation returns the original bytes unchanged (no sequence-number bump / checksum rewrite).
func TestBytesNoOpWhenUnmodified(t *testing.T) {
	orig := buildKeyHive()
	h, err := OpenBytes(append([]byte(nil), orig...))
	if err != nil {
		t.Fatalf("OpenBytes: %v", err)
	}
	out, err := h.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if !bytes.Equal(out, orig) {
		t.Error("Bytes on an unmodified hive changed the image (expected byte-for-byte identity)")
	}

	// After a mutation, Bytes must finalize (bump sequence numbers).
	if err := h.CreateKey("", "X"); err != nil {
		t.Fatalf("CreateKey: %v", err)
	}
	out2, err := h.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	if binary.LittleEndian.Uint32(out2[4:8]) != 2 {
		t.Errorf("primary sequence = %d after mutation, want 2", binary.LittleEndian.Uint32(out2[4:8]))
	}
}

// TestWriteUpdatesTimestampAndHints verifies that mutations stamp LastWrittenTimestamp and
// grow the NK "largest name/data" hint fields.
func TestWriteUpdatesTimestampAndHints(t *testing.T) {
	h := openKeyHive(t)

	if err := h.SetValue("", "AVeryLongValueName", RegBinary, bytes.Repeat([]byte{1}, 300)); err != nil {
		t.Fatalf("SetValue: %v", err)
	}
	root, _ := h.FindKey("")
	if root.LastWrittenTimestamp == 0 {
		t.Error("LastWrittenTimestamp not updated by SetValue")
	}
	if int(root.MaxValueNameLength) < len("AVeryLongValueName") {
		t.Errorf("MaxValueNameLength = %d, want >= %d", root.MaxValueNameLength, len("AVeryLongValueName"))
	}
	if root.MaxValueDataSize < 300 {
		t.Errorf("MaxValueDataSize = %d, want >= 300", root.MaxValueDataSize)
	}

	if err := h.CreateKey("", "ALongSubkeyName"); err != nil {
		t.Fatalf("CreateKey: %v", err)
	}
	root, _ = h.FindKey("")
	if int(root.MaxSubKeyNameLength) < len("ALongSubkeyName") {
		t.Errorf("MaxSubKeyNameLength = %d, want >= %d", root.MaxSubKeyNameLength, len("ALongSubkeyName"))
	}
}

// TestRebuildPreservesUnreadableSubkey verifies that rebuilding a subkey list keeps a
// referenced subkey whose node cannot be parsed, instead of silently dropping it.
func TestRebuildPreservesUnreadableSubkey(t *testing.T) {
	// Build a hive: root -> lh list referencing [validSub, corruptCell].
	a := &hiveAsm{}
	hb := &HiveBin{Signature: hbinSignature, Size: 4096}
	hdr, _ := hb.Marshal()
	a.bins = append(a.bins, hdr...)

	rootName := utf16le("ROOT")
	root := &KeyNode{Signature: nkSignature, Flags: KeyHiveEntry, NumberOfSubKeys: 2,
		ValuesListOffset: nullCellOffset, SecurityOffset: nullCellOffset, ClassNameOffset: nullCellOffset,
		Parent: nullCellOffset, KeyNameLength: uint16(len(rootName)), KeyNameRaw: rootName}
	rootb, _ := root.Marshal()
	rootOff := a.addCell(rootb)
	rootContent := int(rootOff) + 4

	good := &KeyNode{Signature: nkSignature, Flags: KeyCompName, SubKeysListOffset: nullCellOffset,
		ValuesListOffset: nullCellOffset, SecurityOffset: nullCellOffset, ClassNameOffset: nullCellOffset,
		KeyNameLength: uint16(len("Good")), KeyNameRaw: []byte("Good")}
	goodb, _ := good.Marshal()
	goodOff := a.addCell(goodb)

	corruptOff := a.addCell([]byte{0xFF, 0xFF, 0, 0, 0, 0, 0, 0}) // not an "nk" record

	lh := &SubKeyList{Signature: lhSig, NumberOfElements: 2, Elements: make([]byte, 16)}
	binary.LittleEndian.PutUint32(lh.Elements[0:4], goodOff)
	binary.LittleEndian.PutUint32(lh.Elements[8:12], corruptOff)
	lhb, _ := lh.Marshal()
	lhOff := a.addCell(lhb)
	binary.LittleEndian.PutUint32(a.bins[rootContent+28:rootContent+32], lhOff)

	if len(a.bins) < 4096 {
		a.bins = append(a.bins, make([]byte, 4096-len(a.bins))...)
	}
	bb := &BaseBlock{Signature: regfSignature, MajorVersion: 1, MinorVersion: 6, FileFormat: 1,
		RootCellOffset: rootOff, HiveBinsDataSize: uint32(len(a.bins)), ClusteringFactor: 1,
		PrimarySequenceNumber: 1, SecondarySequenceNumber: 1}
	bbb, _ := bb.Marshal()
	hive := append(bbb, a.bins...)

	h, err := OpenBytes(hive)
	if err != nil {
		t.Fatalf("OpenBytes: %v", err)
	}
	entries, _, err := h.readSubkeyEntries(lhOff)
	if err != nil {
		t.Fatalf("readSubkeyEntries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("readSubkeyEntries kept %d of 2 subkeys (unreadable one dropped)", len(entries))
	}

	// A CreateKey rebuild must keep the count consistent (2 existing + 1 new = 3).
	if err := h.CreateKey("", "New"); err != nil {
		t.Fatalf("CreateKey: %v", err)
	}
	root2, _ := h.FindKey("")
	if root2.NumberOfSubKeys != 3 {
		t.Errorf("NumberOfSubKeys = %d after CreateKey, want 3", root2.NumberOfSubKeys)
	}
	if got, _, _ := h.readSubkeyEntries(root2.SubKeysListOffset); len(got) != 3 {
		t.Errorf("rebuilt subkey list has %d entries, want 3", len(got))
	}
}
