package regf

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

// This file is self-contained: it builds its own minimal primary hive and transaction-log
// fixtures rather than reusing the shared sample hive, so transaction-log testing stays
// independent of the other fixtures.
//
// The primary hive is a single 4096-byte hive bin: root key "ROOT" with one inline
// REG_DWORD value "Val". The transaction log carries one HvLE entry whose single dirty
// page replaces the first 4096 bytes of hive-bins data with a version where "Val" differs,
// so replaying it flips the value.

const (
	txlogBasePath = "testdata/txlog_base.hiv"
	txlogLogPath  = "testdata/txlog_base.LOG1"
	txlogPageSize = 4096
)

// buildTxlogHive builds the minimal primary hive with ROOT\Val = dword (inline).
func buildTxlogHive(dword uint32) []byte {
	a := &hiveAsm{}
	hb := &HiveBin{Signature: hbinSignature, Offset: 0, Size: txlogPageSize}
	hdr, _ := hb.Marshal()
	a.bins = append(a.bins, hdr...)

	rootName := utf16le("ROOT")
	root := &KeyNode{
		Signature:         nkSignature,
		Flags:             KeyHiveEntry,
		NumberOfValues:    1,
		ValuesListOffset:  nullCellOffset, // back-patched
		SubKeysListOffset: nullCellOffset,
		SecurityOffset:    nullCellOffset,
		ClassNameOffset:   nullCellOffset,
		Parent:            nullCellOffset,
		KeyNameLength:     uint16(len(rootName)),
		KeyNameRaw:        rootName,
	}
	rootb, _ := root.Marshal()
	rootOff := a.addCell(rootb)
	rootContent := int(rootOff) + 4

	vk := &KeyValue{
		Signature:  vkSignature,
		NameLength: uint16(len("Val")),
		DataSize:   0x80000000 | 4,
		DataOffset: dword,
		DataType:   RegDword,
		Flags:      ValueCompName,
		NameRaw:    []byte("Val"),
	}
	vkb, _ := vk.Marshal()
	vkOff := a.addCell(vkb)

	valueList := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueList, vkOff)
	valueListOff := a.addCell(valueList)

	// ValuesListOffset is at NK+40.
	binary.LittleEndian.PutUint32(a.bins[rootContent+40:rootContent+44], valueListOff)

	if len(a.bins) < txlogPageSize {
		a.bins = append(a.bins, make([]byte, txlogPageSize-len(a.bins))...)
	}

	bb := &BaseBlock{
		Signature:               regfSignature,
		PrimarySequenceNumber:   1,
		SecondarySequenceNumber: 1,
		MajorVersion:            1,
		MinorVersion:            6,
		FileFormat:              1,
		RootCellOffset:          rootOff,
		HiveBinsDataSize:        uint32(len(a.bins)),
		ClusteringFactor:        1,
	}
	bbb, _ := bb.Marshal()
	var cs uint32
	for i := 0; i < 508; i += 4 {
		cs ^= binary.LittleEndian.Uint32(bbb[i : i+4])
	}
	switch cs {
	case 0:
		cs = 1
	case 0xFFFFFFFF:
		cs = 0xFFFFFFFE
	}
	binary.LittleEndian.PutUint32(bbb[508:512], cs)
	return append(bbb, a.bins...)
}

// buildTxlogLog builds a transaction-log file whose single HvLE entry (sequence 1) carries
// one dirty page: the first 4096 bytes of modified's hive-bins data, to be written at
// offset 0 of the primary's hive-bins data.
func buildTxlogLog(base, modified []byte) []byte {
	page := append([]byte(nil), modified[baseBlockSize:baseBlockSize+txlogPageSize]...)
	entry := &LogEntry{
		Signature:        hvleSignature,
		SequenceNumber:   1,
		HiveBinsDataSize: txlogPageSize,
		DirtyPagesCount:  1,
		DirtyPages:       []DirtyPageReference{{Offset: 0, Size: txlogPageSize}},
		PageData:         [][]byte{page},
	}
	body := logEntryHeaderSize + 8 + txlogPageSize
	entry.LogSize = uint32((body + 511) &^ 511) // 512-aligned
	entryBytes, _ := entry.Marshal()

	// 512-byte log header: a copy of the primary base block marked as a log file.
	header := make([]byte, logEntriesOffset)
	copy(header, base[:logEntriesOffset])
	binary.LittleEndian.PutUint32(header[28:32], 1) // FileType = transaction log
	return append(header, entryBytes...)
}

// TestTxlogGoldenFixtures keeps the committed transaction-log fixtures in sync with the
// builders. Regenerate with `go test -run TestTxlogGoldenFixtures -update`.
func TestTxlogGoldenFixtures(t *testing.T) {
	base := buildTxlogHive(0x11111111)
	modified := buildTxlogHive(0x22222222)
	log := buildTxlogLog(base, modified)

	if *update {
		if err := os.WriteFile(txlogBasePath, base, 0o644); err != nil {
			t.Fatalf("write base: %v", err)
		}
		if err := os.WriteFile(txlogLogPath, log, 0o644); err != nil {
			t.Fatalf("write log: %v", err)
		}
		t.Logf("wrote %s (%d) and %s (%d)", txlogBasePath, len(base), txlogLogPath, len(log))
	}
	for path, want := range map[string][]byte{txlogBasePath: base, txlogLogPath: log} {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s (run with -update to create): %v", path, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s out of date; run: go test -run TestTxlogGoldenFixtures -update", path)
		}
	}
}

func TestLogEntryRoundTrip(t *testing.T) {
	orig := LogEntry{
		Signature:        hvleSignature,
		LogSize:          512,
		SequenceNumber:   7,
		HiveBinsDataSize: 4096,
		DirtyPagesCount:  1,
		Hash1:            0x1122334455667788,
		DirtyPages:       []DirtyPageReference{{Offset: 0x40, Size: 8}},
		PageData:         [][]byte{{1, 2, 3, 4, 5, 6, 7, 8}},
	}
	raw, _ := orig.Marshal()
	var got LogEntry
	if _, err := got.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Signature != orig.Signature || got.SequenceNumber != orig.SequenceNumber ||
		got.DirtyPagesCount != 1 || len(got.DirtyPages) != 1 ||
		got.DirtyPages[0] != orig.DirtyPages[0] || !bytes.Equal(got.PageData[0], orig.PageData[0]) {
		t.Errorf("round-trip mismatch:\n orig=%+v\n  got=%+v", orig, got)
	}
}

func dwordValue(t *testing.T, data []byte) uint32 {
	t.Helper()
	h, err := OpenBytes(data)
	if err != nil {
		t.Fatalf("OpenBytes: %v", err)
	}
	root, err := h.FindKey("")
	if err != nil {
		t.Fatalf("FindKey root: %v", err)
	}
	v, err := root.Value("Val")
	if err != nil {
		t.Fatalf("Value(Val): %v", err)
	}
	n, ok := v.Uint32()
	if !ok {
		t.Fatal("Val is not a DWORD")
	}
	return n
}

func TestTransactionLogReplay(t *testing.T) {
	base := buildTxlogHive(0x11111111)
	modified := buildTxlogHive(0x22222222)
	log := buildTxlogLog(base, modified)

	// Before replay the primary holds the original value.
	if got := dwordValue(t, base); got != 0x11111111 {
		t.Fatalf("base ROOT\\Val = 0x%08X, want 0x11111111", got)
	}

	recovered, applied, err := ReplayTransactionLog(base, log)
	if err != nil {
		t.Fatalf("ReplayTransactionLog: %v", err)
	}
	if applied != 1 {
		t.Errorf("applied = %d, want 1", applied)
	}
	if got := dwordValue(t, recovered); got != 0x22222222 {
		t.Errorf("recovered ROOT\\Val = 0x%08X, want 0x22222222", got)
	}

	// The base block's sequence numbers must advance to reflect the applied transaction.
	var rb BaseBlock
	if _, err := rb.Unmarshal(recovered); err != nil {
		t.Fatalf("recovered base block: %v", err)
	}
	if rb.PrimarySequenceNumber != 2 || rb.SecondarySequenceNumber != 2 {
		t.Errorf("recovered sequence numbers = %d/%d, want 2/2", rb.PrimarySequenceNumber, rb.SecondarySequenceNumber)
	}
}

func TestOpenWithLogs(t *testing.T) {
	h, err := OpenWithLogs(txlogBasePath, txlogLogPath)
	if err != nil {
		t.Fatalf("OpenWithLogs: %v", err)
	}
	defer h.Close()
	root, err := h.FindKey("")
	if err != nil {
		t.Fatalf("FindKey: %v", err)
	}
	v, err := root.Value("Val")
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	if n, ok := v.Uint32(); !ok || n != 0x22222222 {
		t.Errorf("OpenWithLogs ROOT\\Val = 0x%08X (ok=%v), want 0x22222222", n, ok)
	}

	// With no logs, OpenWithLogs is equivalent to Open (unrecovered value).
	h2, err := OpenWithLogs(txlogBasePath)
	if err != nil {
		t.Fatalf("OpenWithLogs (no logs): %v", err)
	}
	defer h2.Close()
	root2, _ := h2.FindKey("")
	v2, _ := root2.Value("Val")
	if n, _ := v2.Uint32(); n != 0x11111111 {
		t.Errorf("OpenWithLogs without logs = 0x%08X, want 0x11111111", n)
	}
}
