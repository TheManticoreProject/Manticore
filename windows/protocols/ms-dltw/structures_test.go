package msdltw

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts the
// result is deeply equal to in. This is the wire-shape acceptance gate for the MS-DLTW
// NDR structures in the absence of a live DLT Workstation server.
func roundTrip[T any](t *testing.T, name string, in T) []byte {
	t.Helper()
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}
	var out T
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("%s: round trip mismatch:\n in:  %+v\n out: %+v", name, in, out)
	}
	return raw
}

// sampleGUID is a fixed msdtyp.GUID used across the tests; it exercises every octet slot so
// a byte-order or truncation bug shows up in the round trip.
var sampleGUID = msdtyp.GUID{
	Data1: 0x11223344,
	Data2: 0x5566,
	Data3: 0x7788,
	Data4: [8]byte{0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00},
}

// TestCObjId round-trips an ObjectID and pins its 16-octet wire size — the guard against
// modeling the GUID on windows/guid.GUID, whose trailing uint64 marshals to 24 octets.
func TestCObjId(t *testing.T) {
	raw := roundTrip(t, "CObjId", CObjId{Object: sampleGUID})
	if len(raw) != 16 {
		t.Errorf("CObjId marshaled to %d octets, want 16", len(raw))
	}
}

// TestCVolumeId round-trips a VolumeID and pins its 16-octet wire size.
func TestCVolumeId(t *testing.T) {
	raw := roundTrip(t, "CVolumeId", CVolumeId{Volume: sampleGUID})
	if len(raw) != 16 {
		t.Errorf("CVolumeId marshaled to %d octets, want 16", len(raw))
	}
}

// TestCMachineId round-trips a MachineID (a fixed 16-octet char array) and checks that
// String trims at the NUL terminator.
func TestCMachineId(t *testing.T) {
	m := NewCMachineId("SERVER01")
	raw := roundTrip(t, "CMachineId", m)
	if len(raw) != 16 {
		t.Errorf("CMachineId marshaled to %d octets, want 16", len(raw))
	}
	if got := m.String(); got != "SERVER01" {
		t.Errorf("CMachineId.String() = %q, want %q", got, "SERVER01")
	}
}

// TestCMachineIdFull checks the boundary where the name fills the field without a NUL:
// NewCMachineId keeps at most 15 octets so a terminator always remains.
func TestCMachineIdFull(t *testing.T) {
	m := NewCMachineId("0123456789ABCDEFGH") // longer than 15
	if got := m.String(); got != "0123456789ABCDE" {
		t.Errorf("CMachineId.String() = %q, want %q", got, "0123456789ABCDE")
	}
	if m.Machine[15] != 0 {
		t.Errorf("CMachineId last octet = 0x%02x, want NUL terminator", m.Machine[15])
	}
}

// TestCDomainRelativeObjId round-trips the {VolumeID, ObjectID} pair and pins its 32-octet
// wire size (two inline 16-octet GUIDs, in declaration order).
func TestCDomainRelativeObjId(t *testing.T) {
	droid := CDomainRelativeObjId{
		Volume: CVolumeId{Volume: sampleGUID},
		Object: CObjId{Object: msdtyp.GUID{Data1: 0xdeadbeef, Data2: 0x0102, Data3: 0x0304, Data4: [8]byte{1, 2, 3, 4, 5, 6, 7, 8}}},
	}
	raw := roundTrip(t, "CDomainRelativeObjId", droid)
	if len(raw) != 32 {
		t.Errorf("CDomainRelativeObjId marshaled to %d octets, want 32", len(raw))
	}
}
