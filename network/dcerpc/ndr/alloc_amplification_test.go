package ndr

import (
	"encoding/binary"
	"runtime"
	"testing"
)

// amplifiedElem is far larger in memory than the minimum wire size of an element
// count, which is what lets a short stub request a disproportionate allocation.
type amplifiedElem struct {
	A [32]uint64
	S string
}

// TestUnmarshal_ArrayCountAllocationAmplificationBounded verifies that a conformant
// array whose count fits the remaining bytes, but whose elements are much larger in
// memory than on the wire, does not allocate the whole backing array before the
// elements are actually read. A 256 KiB stub claiming one element per remaining byte
// used to allocate count*sizeof(elem) (~70 MiB) before failing.
func TestUnmarshal_ArrayCountAllocationAmplificationBounded(t *testing.T) {
	type rec struct {
		P []amplifiedElem `ndr:"unique,conformant"`
	}
	const dataLen = 256 << 10
	stub := make([]byte, 8+dataLen)
	binary.LittleEndian.PutUint32(stub[0:4], 0x00020000)      // referent id
	binary.LittleEndian.PutUint32(stub[4:8], uint32(dataLen)) // maximum_count = remaining bytes

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	var out rec
	err := Unmarshal(stub, &out)
	runtime.ReadMemStats(&after)

	if err == nil {
		t.Fatal("Unmarshal of a truncated oversized array: error = nil, want non-nil")
	}
	const limit = 16 << 20
	if got := after.TotalAlloc - before.TotalAlloc; got > limit {
		t.Errorf("Unmarshal allocated %d bytes for a %d-byte stub, want <= %d", got, len(stub), limit)
	}
}

// TestUnmarshal_LargeArrayChunkedDecodeRoundTrip verifies an array large enough to be
// decoded in chunks, with deferred element referents, round-trips unchanged.
func TestUnmarshal_LargeArrayChunkedDecodeRoundTrip(t *testing.T) {
	type elem struct {
		A [16]uint64
		P *uint32 `ndr:"unique"`
	}
	type rec struct {
		P []elem `ndr:"unique,conformant"`
	}
	const n = 3 * (1 << 20) / 128 // several chunks
	in := &rec{P: make([]elem, n)}
	for i := range in.P {
		in.P[i].A[0] = uint64(i)
		if i%3 == 0 {
			v := uint32(i)
			in.P[i].P = &v
		}
	}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out rec
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.P) != n {
		t.Fatalf("len = %d, want %d", len(out.P), n)
	}
	for i := range out.P {
		if out.P[i].A[0] != uint64(i) {
			t.Fatalf("element %d: A[0] = %d", i, out.P[i].A[0])
		}
		if i%3 == 0 {
			if out.P[i].P == nil || *out.P[i].P != uint32(i) {
				t.Fatalf("element %d: P = %v, want %d", i, out.P[i].P, i)
			}
		} else if out.P[i].P != nil {
			t.Fatalf("element %d: P = %v, want nil", i, *out.P[i].P)
		}
	}
}
