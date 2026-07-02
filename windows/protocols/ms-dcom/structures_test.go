package msdcom

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts the
// result is deeply equal to in. This is the wire-shape acceptance gate for the MS-DCOM
// IObjectExporter (IOXIDResolver) NDR structures in the absence of a live object resolver.
func roundTrip[T any](t *testing.T, name string, in T) {
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
}

// TestCOMVERSION exercises the fixed two-uint16 version structure.
func TestCOMVERSION(t *testing.T) {
	roundTrip(t, "COMVERSION", COMVERSION{MajorVersion: 5, MinorVersion: 7})
}

// TestDUALSTRINGARRAY exercises the bare inline conformant array (aStringArray) whose
// maximum_count is transmitted in place with no referent id. The buffer is the wNumEntries
// unsigned shorts holding the string bindings followed by the security bindings.
func TestDUALSTRINGARRAY(t *testing.T) {
	roundTrip(t, "DUALSTRINGARRAY/populated", DUALSTRINGARRAY{
		WNumEntries:     6,
		WSecurityOffset: 4,
		// two STRINGBINDING shorts + null, then a SECURITYBINDING short + null.
		AStringArray: []uint16{0x0007, 0x4142, 0x0000, 0x000a, 0x0000, 0x0000},
	})
	roundTrip(t, "DUALSTRINGARRAY/empty", DUALSTRINGARRAY{
		WNumEntries:     0,
		WSecurityOffset: 0,
		AStringArray:    []uint16{},
	})
}

// TestScalarTypedefs exercises the unsigned hyper aliases OXID/OID/SETID. NDR marshals
// top-level structs only, so they are wrapped (as they appear on the wire — inline fields
// of a request/response struct).
func TestScalarTypedefs(t *testing.T) {
	type scalars struct {
		Oxid  OXID
		Oid   OID
		SetId SETID
	}
	roundTrip(t, "scalars", scalars{
		Oxid:  0x1122334455667788,
		Oid:   0x0102030405060708,
		SetId: 0xfedcba9876543210,
	})
}

// TestIPID exercises the GUID-shaped IPID, which must marshal to exactly 16 octets.
func TestIPID(t *testing.T) {
	ipid := IPID{Data1: 0x00000131, Data2: 0x0000, Data3: 0x0000,
		Data4: [8]byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	roundTrip(t, "IPID", ipid)

	raw, err := ndr.Marshal(&ipid)
	if err != nil {
		t.Fatalf("IPID: Marshal: %v", err)
	}
	if len(raw) != 16 {
		t.Fatalf("IPID marshaled to %d octets, want 16", len(raw))
	}
	// The IPID must render identically through its own helper and the shared dtyp.GUID.
	if got, want := ipid.String(), dtyp.GUID(ipid).String(); got != want {
		t.Fatalf("IPID/dtyp.GUID string mismatch: %s vs %s", got, want)
	}
}
