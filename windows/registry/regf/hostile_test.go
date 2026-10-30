package regf

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"testing"
)

// The tests in this file feed hostile 32-bit counts, sizes and offsets to the parser. On a
// 32-bit build (GOARCH=386) the unfixed code wrapped them into negative ints and panicked;
// on 64-bit builds several of them forced multi-gigabyte allocations. Each test asserts an
// error (or a clamped result) and, where an allocation was the problem, that the call
// allocates only a small amount of memory.

// hostileHive returns a minimal 8 KiB in-memory hive image whose first hive-bins cell (at
// offset 0) is an allocated 16-byte cell.
func hostileHive() *Hive {
	d := make([]byte, 2*baseBlockSize)
	binary.LittleEndian.PutUint32(d[baseBlockSize:baseBlockSize+4], uint32(0xFFFFFFF0)) // -16
	return &Hive{data: d}
}

// maxHostileAlloc is the allocation ceiling for a single call on a hostile input; the
// unfixed code allocated gigabytes for the same inputs.
const maxHostileAlloc = 64 << 20

// allocatedBy returns how many bytes f allocated.
func allocatedBy(f func()) uint64 {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

func TestSecurityKeyHostileDescriptorSize(t *testing.T) {
	d := make([]byte, skMinSize+8)
	binary.LittleEndian.PutUint16(d[0:2], skSignature)
	binary.LittleEndian.PutUint32(d[16:20], 0xFFFFFFF0)
	sk := NewSecurityKey()
	n, err := sk.Unmarshal(d)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != len(d) || len(sk.SecurityDescriptor) != 8 {
		t.Errorf("consumed %d, descriptor %d bytes; want %d and 8 (clamped to the buffer)", n, len(sk.SecurityDescriptor), len(d))
	}
}

func TestLogEntryHostileCounts(t *testing.T) {
	for _, tc := range []struct {
		name            string
		count, pageSize uint32
	}{
		{"dirty-pages-count", 0x20000000, 0},
		{"dirty-pages-count-max", 0xFFFFFFFF, 0},
		{"page-size", 1, 0xFFFFFFF0},
		{"page-size-high-bit", 1, 0x80000000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := make([]byte, 64)
			binary.LittleEndian.PutUint32(d[0:4], hvleSignature)
			binary.LittleEndian.PutUint32(d[20:24], tc.count)
			binary.LittleEndian.PutUint32(d[44:48], tc.pageSize)
			if _, err := NewLogEntry().Unmarshal(d); err == nil {
				t.Errorf("Unmarshal accepted DirtyPagesCount=0x%X page size=0x%X", tc.count, tc.pageSize)
			}
		})
	}
}

func TestLogEntryHostileLogSize(t *testing.T) {
	d := make([]byte, 64)
	binary.LittleEndian.PutUint32(d[0:4], hvleSignature)
	binary.LittleEndian.PutUint32(d[4:8], 0xFFFFFFF0) // LogSize
	n, err := NewLogEntry().Unmarshal(d)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if n != len(d) {
		t.Errorf("consumed %d, want %d (an oversized LogSize ends at the buffer)", n, len(d))
	}
	var tl TransactionLog
	if err := tl.Unmarshal(append(make([]byte, logEntriesOffset), d...)); err != nil || len(tl.Entries) != 1 {
		t.Errorf("TransactionLog.Unmarshal = %d entries, %v; want 1 entry", len(tl.Entries), err)
	}
}

func TestHostileCellOffsets(t *testing.T) {
	h := hostileHive()
	for _, off := range []uint32{0x7FFFFFFF, 0x80000000, 0xFFFFEFFC, 0xFFFFFFFE} {
		if _, err := h.readCellRaw(off); err == nil {
			t.Errorf("readCellRaw(0x%X) succeeded", off)
		}
		if _, err := h.readCellData(off, 4); err == nil {
			t.Errorf("readCellData(0x%X, 4) succeeded", off)
		}
		if _, ok := h.cellSizeAt(off); ok {
			t.Errorf("cellSizeAt(0x%X) succeeded", off)
		}
		if got := h.valueListOffsets(off, 1); got != nil {
			t.Errorf("valueListOffsets(0x%X) = %v, want nil", off, got)
		}
		k := &KeyNode{hive: h, SecurityOffset: off}
		if _, err := k.SecurityDescriptor(); err == nil {
			t.Errorf("SecurityDescriptor with SecurityOffset 0x%X succeeded", off)
		}
		h.bumpSKRefCount(off, 1) // must not panic
	}
}

func TestHostileCellSize(t *testing.T) {
	h := hostileHive()
	// A free cell claiming the largest positive size must be clamped to the image.
	binary.LittleEndian.PutUint32(h.data[baseBlockSize:baseBlockSize+4], 0x7FFFFFFF)
	cell, err := h.readCellRaw(0)
	if err != nil {
		t.Fatalf("readCellRaw: %v", err)
	}
	if len(cell) != baseBlockSize-4 {
		t.Errorf("cell length %d, want %d", len(cell), baseBlockSize-4)
	}
}

func TestKeyValueHostileDataSize(t *testing.T) {
	v := &KeyValue{hive: hostileHive(), DataOffset: 0, DataSize: 0x7FFFFFF0}
	if _, err := v.Data(); err == nil {
		t.Error("Data with DataSize 0x7FFFFFF0 in an 8 KiB hive succeeded")
	}
}

func TestReadValueListHostileCount(t *testing.T) {
	h := hostileHive()
	k := &KeyNode{hive: h, ValuesListOffset: 0, NumberOfValues: 0xFFFFFFFF}
	if n := allocatedBy(func() { _, _ = k.Values() }); n > maxHostileAlloc {
		t.Errorf("Values with NumberOfValues 0xFFFFFFFF allocated %d bytes", n)
	}
	var offsets []uint32
	if n := allocatedBy(func() { offsets = h.valueListOffsets(0, 0xFFFFFFFF) }); n > maxHostileAlloc {
		t.Errorf("valueListOffsets with count 0xFFFFFFFF allocated %d bytes", n)
	}
	if want := (len(h.data) - baseBlockSize - 4) / 4; len(offsets) != want {
		t.Errorf("valueListOffsets returned %d offsets, want %d (bounded by the image)", len(offsets), want)
	}
}

func TestReadBigDataHostileSize(t *testing.T) {
	h := hostileHive()
	// db record at offset 0: one segment whose list is the db cell itself.
	binary.LittleEndian.PutUint16(h.data[baseBlockSize+4:baseBlockSize+6], dbSignature)
	binary.LittleEndian.PutUint16(h.data[baseBlockSize+6:baseBlockSize+8], 1)
	v := &KeyValue{hive: h, DataOffset: 0, DataSize: 0x7FFFFFF0}
	var data []byte
	if n := allocatedBy(func() { data, _ = v.Data() }); n > maxHostileAlloc {
		t.Errorf("Data with DataSize 0x7FFFFFF0 allocated %d bytes from an 8 KiB hive", n)
	}
	if len(data) > len(h.data) {
		t.Errorf("Data returned %d bytes from an %d-byte hive", len(data), len(h.data))
	}
}

// hvleLog builds a transaction log with a single HvLE entry that continues base's sequence
// and writes page at offset, declaring hiveBinsDataSize.
func hvleLog(base []byte, offset, hiveBinsDataSize uint32, page []byte) []byte {
	var bb BaseBlock
	bb.Unmarshal(base)
	entry := &LogEntry{
		Signature:        hvleSignature,
		SequenceNumber:   bb.SecondarySequenceNumber,
		HiveBinsDataSize: hiveBinsDataSize,
		DirtyPagesCount:  1,
		DirtyPages:       []DirtyPageReference{{Offset: offset, Size: uint32(len(page))}},
		PageData:         [][]byte{page},
	}
	entry.LogSize = uint32((logEntryHeaderSize + 8 + len(page) + 511) &^ 511)
	entryBytes, _ := entry.Marshal()
	return append(make([]byte, logEntriesOffset), entryBytes...)
}

func TestReplayHostilePageOffset(t *testing.T) {
	base := buildTxlogHive(0x11111111)
	for _, tc := range []struct {
		name                     string
		offset, hiveBinsDataSize uint32
	}{
		{"beyond-declared-size", 0xFFFFE000, 0x1000},
		{"declared-size-too-large", 0xFFFFE000, 0xFFFFF000},
		{"high-bit", 0x80000000, 0x80001000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			log := hvleLog(base, tc.offset, tc.hiveBinsDataSize, make([]byte, 8))
			var (
				recovered []byte
				applied   int
				err       error
			)
			n := allocatedBy(func() { recovered, applied, err = ReplayTransactionLog(base, log) })
			if err != nil {
				t.Fatalf("ReplayTransactionLog: %v (a bogus entry must end recovery, not fail it)", err)
			}
			if applied != 0 || !bytes.Equal(recovered, base) {
				t.Errorf("applied %d entries (image %d bytes); want the bogus entry skipped and the image unchanged (%d bytes)",
					applied, len(recovered), len(base))
			}
			if n > maxHostileAlloc {
				t.Errorf("ReplayTransactionLog allocated %d bytes", n)
			}
		})
	}
}

// TestReplayGrowsHive checks that a well-formed entry appending a new bin (fully logged, as
// Windows logs every new bin) still grows the image.
func TestReplayGrowsHive(t *testing.T) {
	base := buildTxlogHive(0x11111111)
	page := make([]byte, 4096)
	copy(page, "hbin")
	var bb BaseBlock
	bb.Unmarshal(base)
	end := bb.HiveBinsDataSize
	recovered, applied, err := ReplayTransactionLog(base, hvleLog(base, end, end+4096, page))
	if err != nil {
		t.Fatalf("ReplayTransactionLog: %v", err)
	}
	if applied != 1 || len(recovered) != baseBlockSize+int(end)+4096 {
		t.Errorf("applied %d, image %d bytes; want 1 and %d", applied, len(recovered), baseBlockSize+int(end)+4096)
	}
}

func TestReplayDirtHostileBitmap(t *testing.T) {
	base := buildTxlogHive(0x11111111)
	binary.LittleEndian.PutUint32(base[40:44], 0xFFFFF000) // HiveBinsDataSize
	bitmapLen := 0xFFFFF000 / dirtPageSize / 8
	log := make([]byte, dirtDataOffset+bitmapLen+dirtPageSize)
	copy(log[logEntriesOffset:], "DIRT")
	log[logEntriesOffset+4+bitmapLen-1] = 0x80 // only the last 512-byte block is dirty
	var (
		recovered []byte
		applied   int
		err       error
	)
	n := allocatedBy(func() { recovered, applied, err = ReplayTransactionLog(base, log) })
	if err != nil {
		t.Fatalf("ReplayTransactionLog: %v (a bogus block must end recovery, not fail it)", err)
	}
	if applied != 0 || !bytes.Equal(recovered, base) {
		t.Errorf("applied %d blocks (image %d bytes); want the block ~4 GiB past the end skipped and the image unchanged (%d bytes)",
			applied, len(recovered), len(base))
	}
	if n > maxHostileAlloc {
		t.Errorf("ReplayTransactionLog allocated %d bytes", n)
	}
}

// TestReplayStopsAtBogusEntry checks that a bogus entry ends recovery while keeping the
// well-formed entries that precede it.
func TestReplayStopsAtBogusEntry(t *testing.T) {
	base := buildTxlogHive(0x11111111)
	log := buildTxlogLog(base, buildTxlogHive(0x22222222)) // sequence 1: valid
	bogus := &LogEntry{
		Signature:        hvleSignature,
		SequenceNumber:   2,
		HiveBinsDataSize: 0xFFFFF000,
		DirtyPagesCount:  1,
		DirtyPages:       []DirtyPageReference{{Offset: 0xFFFFE000, Size: 8}},
		PageData:         [][]byte{make([]byte, 8)},
	}
	bogus.LogSize = 512
	bogusBytes, _ := bogus.Marshal()
	log = append(log, bogusBytes...)

	recovered, applied, err := ReplayTransactionLog(base, log)
	if err != nil {
		t.Fatalf("ReplayTransactionLog: %v", err)
	}
	if applied != 1 {
		t.Errorf("applied = %d, want 1 (the entry preceding the bogus one)", applied)
	}
	if len(recovered) != len(base) {
		t.Errorf("image %d bytes, want %d", len(recovered), len(base))
	}
	if got := dwordValue(t, recovered); got != 0x22222222 {
		t.Errorf("recovered ROOT\\Val = 0x%08X, want 0x22222222", got)
	}
}
