package ndr

import (
	"reflect"
	"testing"
)

// Fixed (non-conformant) arrays of non-byte elements ([C706] 14.3.4) marshal their
// elements in place with no count prefix. These guard the marshalInlineValue /
// unmarshalInlineValue Array cases for scalar, struct, and pointer element kinds
// (MS-DNSP's DwReserveArray[N], DnsAddrUserDword[8], ZoneQueryStats[32], and the
// pExtensions[6] string-pointer arrays all depend on this).

func fixedRT[T any](t *testing.T, name string, in T, wantLen int) {
	t.Helper()
	raw, err := Marshal(&in)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}
	if wantLen >= 0 && len(raw) != wantLen {
		t.Errorf("%s: marshaled %d octets, want %d", name, len(raw), wantLen)
	}
	var out T
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("%s: round trip mismatch:\n in:  %+v\n out: %+v", name, in, out)
	}
}

func TestFixedArrayDword(t *testing.T) {
	type s struct {
		A [4]DWORD
	}
	fixedRT(t, "fixed[4]DWORD", s{A: [4]DWORD{1, 2, 3, 4}}, 16)
}

func TestFixedArrayWord(t *testing.T) {
	type s struct {
		A [3]uint16
	}
	fixedRT(t, "fixed[3]uint16", s{A: [3]uint16{0x1111, 0x2222, 0x3333}}, 6)
}

func TestFixedArrayStruct(t *testing.T) {
	type inner struct {
		X DWORD
		Y uint64
	}
	type s struct {
		A [2]inner
	}
	// Each inner aligns its uint64 to 8: {DWORD, pad4, uint64} = 16 octets, so 2 => 32.
	fixedRT(t, "fixed[2]struct", s{A: [2]inner{{1, 2}, {3, 4}}}, 32)
}

func TestFixedArrayUniquePtr(t *testing.T) {
	type s struct {
		A [3]*DWORD `ndr:"elem=unique"`
	}
	v1, v3 := DWORD(0xaa), DWORD(0xcc)
	// 3 referent ids (one is NULL) then the two non-nil referent bodies.
	fixedRT(t, "fixed[3]*DWORD", s{A: [3]*DWORD{&v1, nil, &v3}}, -1)
}
