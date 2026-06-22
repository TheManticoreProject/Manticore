package regf

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// le32 returns the little-endian bytes of v (for REG_DWORD values).
func le32(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

// openWritable returns a fresh writable hive (ROOT with one inline value "Val").
func openWritable(t *testing.T) *Hive {
	t.Helper()
	h, err := OpenBytes(buildTxlogHive(0x11111111))
	if err != nil {
		t.Fatalf("OpenBytes: %v", err)
	}
	return h
}

func mustGet(t *testing.T, h *Hive, key, name string) (uint32, []byte) {
	t.Helper()
	typ, data, err := h.GetValue(key, name)
	if err != nil {
		t.Fatalf("GetValue(%q,%q): %v", key, name, err)
	}
	return typ, data
}

func TestSetValueNewInlineAndExternal(t *testing.T) {
	h := openWritable(t)

	// Inline (<= 4 bytes) REG_DWORD.
	if err := h.SetValue("", "Dw", RegDword, le32(0xAABBCCDD)); err != nil {
		t.Fatalf("SetValue Dw: %v", err)
	}
	// External (> 4 bytes) REG_BINARY.
	bin := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if err := h.SetValue("", "Bin", RegBinary, bin); err != nil {
		t.Fatalf("SetValue Bin: %v", err)
	}

	typ, data := mustGet(t, h, "", "Dw")
	if typ != RegDword || !bytes.Equal(data, le32(0xAABBCCDD)) {
		t.Errorf("Dw = (%d,% x), want REG_DWORD DD CC BB AA", typ, data)
	}
	typ, data = mustGet(t, h, "", "Bin")
	if typ != RegBinary || !bytes.Equal(data, bin) {
		t.Errorf("Bin = (%d,% x), want REG_BINARY %x", typ, data, bin)
	}

	// The original value must survive, and the count must reflect the additions.
	if _, d := mustGet(t, h, "", "Val"); binary.LittleEndian.Uint32(d) != 0x11111111 {
		t.Errorf("original Val changed: % x", d)
	}
	names, _ := h.EnumValues("")
	if len(names) != 3 {
		t.Errorf("value count = %d (%v), want 3", len(names), names)
	}
}

func TestSetValueReplace(t *testing.T) {
	h := openWritable(t)
	if err := h.SetValue("", "Val", RegDword, le32(0x99887766)); err != nil {
		t.Fatalf("SetValue replace: %v", err)
	}
	if _, d := mustGet(t, h, "", "Val"); binary.LittleEndian.Uint32(d) != 0x99887766 {
		t.Errorf("replaced Val = % x, want 66 77 88 99", d)
	}
	if names, _ := h.EnumValues(""); len(names) != 1 {
		t.Errorf("value count = %d, want 1 (replace must not add)", len(names))
	}
}

func TestDeleteValue(t *testing.T) {
	h := openWritable(t)
	if err := h.SetValue("", "Extra", RegSz, []byte{0x68, 0, 0x69, 0, 0, 0}); err != nil {
		t.Fatalf("SetValue Extra: %v", err)
	}
	if err := h.DeleteValue("", "Val"); err != nil {
		t.Fatalf("DeleteValue Val: %v", err)
	}
	if _, _, err := h.GetValue("", "Val"); err == nil {
		t.Error("Val still present after delete")
	}
	names, _ := h.EnumValues("")
	if len(names) != 1 || names[0] != "Extra" {
		t.Errorf("after delete, values = %v, want [Extra]", names)
	}
	if err := h.DeleteValue("", "Nope"); err == nil {
		t.Error("DeleteValue of a missing value returned nil error")
	}
}

func TestWritePersistsThroughBytes(t *testing.T) {
	h := openWritable(t)
	if err := h.SetValue("", "Persisted", RegBinary, bytes.Repeat([]byte{0xAB}, 64)); err != nil {
		t.Fatalf("SetValue: %v", err)
	}
	raw, err := h.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	h2, err := OpenBytes(raw)
	if err != nil {
		t.Fatalf("re-open: %v", err)
	}
	typ, data := mustGet(t, h2, "", "Persisted")
	if typ != RegBinary || !bytes.Equal(data, bytes.Repeat([]byte{0xAB}, 64)) {
		t.Errorf("persisted value mismatch: type %d, len %d", typ, len(data))
	}
	// finalize must bump the sequence numbers.
	if h2.BaseBlock().PrimarySequenceNumber != 2 {
		t.Errorf("sequence number = %d, want 2 after one finalize", h2.BaseBlock().PrimarySequenceNumber)
	}
}

func TestFreeCellReuseAndGrowth(t *testing.T) {
	h := openWritable(t)
	// Delete then re-add: the new value should be allocatable (free-cell reuse path).
	if err := h.DeleteValue("", "Val"); err != nil {
		t.Fatalf("DeleteValue: %v", err)
	}
	// Force hive-bin growth with several sizable values.
	for i := 0; i < 8; i++ {
		name := "Big" + string(rune('A'+i))
		if err := h.SetValue("", name, RegBinary, bytes.Repeat([]byte{byte(i)}, 2048)); err != nil {
			t.Fatalf("SetValue %s: %v", name, err)
		}
	}
	raw, err := h.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}
	h2, err := OpenBytes(raw)
	if err != nil {
		t.Fatalf("re-open after growth: %v", err)
	}
	names, err := h2.EnumValues("")
	if err != nil {
		t.Fatalf("EnumValues: %v", err)
	}
	if len(names) != 8 {
		t.Errorf("after growth, %d values, want 8: %v", len(names), names)
	}
	// Spot-check content survived across the grown bins.
	_, data := mustGet(t, h2, "", "BigC")
	if !bytes.Equal(data, bytes.Repeat([]byte{2}, 2048)) {
		t.Errorf("BigC content corrupted after growth (len %d)", len(data))
	}
}
