package ese

import (
	"encoding/binary"
	"runtime"
	"strings"
	"testing"
)

// maxHostileAlloc is the allocation ceiling for a single call on a hostile input; the
// unfixed code allocated gigabytes (or panicked on 32-bit builds) for the same inputs.
const maxHostileAlloc = 64 << 20

// allocatedBy returns how many bytes f allocated.
func allocatedBy(f func()) uint64 {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

// TestOpenRejectsInvalidPageSize checks that only the page sizes ESE supports are
// accepted; 0x80000000 used to panic in make on 32-bit builds and 0xFFFFFFFE made every
// page read allocate ~4 GiB.
func TestOpenRejectsInvalidPageSize(t *testing.T) {
	for _, ps := range []uint32{0, 2, 512, 1024, 12288, 65536, 0x80000000, 0xFFFFFFFE} {
		hdr := make([]byte, 4096)
		binary.LittleEndian.PutUint32(hdr[4:8], eseSignature)
		binary.LittleEndian.PutUint32(hdr[236:240], ps)
		var err error
		n := allocatedBy(func() { _, err = OpenBytes(hdr) })
		if err == nil {
			t.Errorf("OpenBytes accepted page size 0x%X", ps)
		}
		if n > maxHostileAlloc {
			t.Errorf("OpenBytes with page size 0x%X allocated %d bytes", ps, n)
		}
	}

	// Every page size ESE supports passes the page-size check (the truncated
	// header may still fail later for other reasons).
	for _, ps := range []uint32{2048, 4096, 8192, 16384, 32768} {
		hdr := make([]byte, 4096)
		binary.LittleEndian.PutUint32(hdr[4:8], eseSignature)
		binary.LittleEndian.PutUint32(hdr[236:240], ps)
		if _, err := OpenBytes(hdr); err != nil && strings.Contains(err.Error(), "invalid page size") {
			t.Errorf("OpenBytes rejected the valid page size %d: %v", ps, err)
		}
	}
}

// TestResolveLongValueHostileOffset checks that a long-value segment whose key claims an
// offset far past the data actually stored is rejected instead of allocating up to 4 GiB
// (or slicing with a negative offset on 32-bit builds).
func TestResolveLongValueHostileOffset(t *testing.T) {
	lid := []byte{1, 0, 0, 0}
	for _, off := range []uint32{0x00001000, 0x7FFFFF00, 0x80000000, 0xFFFFFF00} {
		key := []byte{0, 0, 0, 1, 0, 0, 0, 0}
		binary.BigEndian.PutUint32(key[4:], off)
		tb := &Table{lvLoaded: true, lvLeaves: []lvEntry{{key: key, data: []byte{1, 2, 3}}}}
		var err error
		n := allocatedBy(func() { _, err = tb.resolveLongValue(lid) })
		if err == nil {
			t.Errorf("resolveLongValue accepted a segment at offset 0x%X", off)
		}
		if n > maxHostileAlloc {
			t.Errorf("resolveLongValue with segment offset 0x%X allocated %d bytes", off, n)
		}
	}
}
