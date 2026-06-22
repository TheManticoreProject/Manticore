package regf

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

const dirtLogPath = "testdata/dirt_base.LOG1"

// buildDirtLog builds a legacy DIRT transaction log whose single dirty block is the first
// 512 bytes of modified's hive-bins data, to be written at offset 0 of the primary's
// hive-bins data. It reuses the txlog primary (HiveBinsDataSize == 4096, so a 1-byte
// bitmap covers the one bin).
func buildDirtLog(base, modified []byte) []byte {
	out := make([]byte, dirtDataOffset+dirtPageSize)
	copy(out, base[:logEntriesOffset])           // 512-byte log header (REGF base block copy)
	binary.LittleEndian.PutUint32(out[28:32], 1) // FileType = transaction log
	copy(out[logEntriesOffset:logEntriesOffset+4], []byte("DIRT"))
	out[logEntriesOffset+4] = 0x01 // dirty bitmap: bit 0 set (first 512-byte block)
	copy(out[dirtDataOffset:dirtDataOffset+dirtPageSize], modified[baseBlockSize:baseBlockSize+dirtPageSize])
	return out
}

// TestDirtGoldenFixture keeps the committed DIRT log fixture in sync with the builder.
// Regenerate with `go test -run TestDirtGoldenFixture -update`.
func TestDirtGoldenFixture(t *testing.T) {
	log := buildDirtLog(buildTxlogHive(0x11111111), buildTxlogHive(0x22222222))
	if *update {
		if err := os.WriteFile(dirtLogPath, log, 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
		t.Logf("wrote %s (%d bytes)", dirtLogPath, len(log))
	}
	got, err := os.ReadFile(dirtLogPath)
	if err != nil {
		t.Fatalf("read %s (run with -update to create): %v", dirtLogPath, err)
	}
	if !bytes.Equal(got, log) {
		t.Errorf("%s out of date; run: go test -run TestDirtGoldenFixture -update", dirtLogPath)
	}
}

func TestDirtLogReplay(t *testing.T) {
	base := buildTxlogHive(0x11111111)
	log := buildDirtLog(base, buildTxlogHive(0x22222222))

	if got := dwordValue(t, base); got != 0x11111111 {
		t.Fatalf("base ROOT\\Val = 0x%08X, want 0x11111111", got)
	}
	recovered, applied, err := ReplayTransactionLog(base, log)
	if err != nil {
		t.Fatalf("ReplayTransactionLog (DIRT): %v", err)
	}
	if applied != 1 {
		t.Errorf("applied = %d, want 1", applied)
	}
	if got := dwordValue(t, recovered); got != 0x22222222 {
		t.Errorf("recovered ROOT\\Val = 0x%08X, want 0x22222222", got)
	}
}

func TestOpenWithDirtLog(t *testing.T) {
	h, err := OpenWithLogs(txlogBasePath, dirtLogPath)
	if err != nil {
		t.Fatalf("OpenWithLogs (DIRT): %v", err)
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
		t.Errorf("OpenWithLogs(DIRT) ROOT\\Val = 0x%08X (ok=%v), want 0x22222222", n, ok)
	}
}
