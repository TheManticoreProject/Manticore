package msmqmq

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals v under both transfer syntaxes, unmarshals into a fresh value of the
// same type, and asserts the result equals v.
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	for _, s := range []ndr.Syntax{ndr.NDR20, ndr.NDR64} {
		raw, err := ndr.MarshalAs(&in, s)
		if err != nil {
			t.Fatalf("%s %s marshal: %v", name, s, err)
		}
		var out T
		if err := ndr.UnmarshalAs(raw, &out, s); err != nil {
			t.Fatalf("%s %s unmarshal: %v", name, s, err)
		}
		if !reflect.DeepEqual(in, out) {
			t.Errorf("%s %s round-trip:\n got %+v\nwant %+v", name, s, out, in)
		}
	}
}

func wstr(s string) *ndr.WSTR { w := ndr.WSTR(s); return &w }

func TestCountedArraysRoundTrip(t *testing.T) {
	roundTrip(t, "BLOB", BLOB{CbSize: 3, PBlobData: []uint8{1, 2, 3}})
	roundTrip(t, "CAUB", CAUB{CElems: 2, PElems: []uint8{0xAA, 0xBB}})
	roundTrip(t, "CAUI", CAUI{CElems: 2, PElems: []uint16{0x1111, 0x2222}})
	roundTrip(t, "CAL", CAL{CElems: 2, PElems: []int32{-5, 7}})
	roundTrip(t, "CAUL", CAUL{CElems: 2, PElems: []ndr.DWORD{100, 200}})
	roundTrip(t, "CAUH", CAUH{CElems: 1, PElems: []dtyp.ULARGE_INTEGER{0x1122334455667788}})
	roundTrip(t, "CACLSID", CACLSID{CElems: 1, PElems: []dtyp.GUID{{Data1: 0xDEAD, Data2: 1, Data3: 2, Data4: [8]byte{3, 4, 5, 6, 7, 8, 9, 10}}}})
	roundTrip(t, "CALPWSTR", CALPWSTR{CElems: 2, PElems: []*ndr.WSTR{wstr("alpha"), wstr("beta")}})
}

func TestPropVariantScalarArms(t *testing.T) {
	var p PROPVARIANT

	p = PROPVARIANT{Value: PropVariantUnion{BVal: 0xAB}}
	p.SetVt(VT_UI1)
	roundTrip(t, "VT_UI1", p)

	p = PROPVARIANT{Value: PropVariantUnion{UlVal: 0xCAFEBABE}}
	p.SetVt(VT_UI4)
	roundTrip(t, "VT_UI4", p)

	p = PROPVARIANT{Value: PropVariantUnion{UhVal: 0x1122334455667788}}
	p.SetVt(VT_UI8)
	roundTrip(t, "VT_UI8", p)

	p = PROPVARIANT{Value: PropVariantUnion{LVal: -42}}
	p.SetVt(VT_I4)
	roundTrip(t, "VT_I4", p)
}

func TestPropVariantPointerArms(t *testing.T) {
	var p PROPVARIANT

	p = PROPVARIANT{Value: PropVariantUnion{PwszVal: wstr("queue-path")}}
	p.SetVt(VT_LPWSTR)
	roundTrip(t, "VT_LPWSTR", p)

	p = PROPVARIANT{Value: PropVariantUnion{Puuid: &dtyp.GUID{Data1: 0x1234, Data4: [8]byte{1, 2, 3, 4, 5, 6, 7, 8}}}}
	p.SetVt(VT_CLSID)
	roundTrip(t, "VT_CLSID", p)
}

func TestPropVariantVectorAndBlobArms(t *testing.T) {
	var p PROPVARIANT

	p = PROPVARIANT{Value: PropVariantUnion{Blob: BLOB{CbSize: 4, PBlobData: []uint8{9, 8, 7, 6}}}}
	p.SetVt(VT_BLOB)
	roundTrip(t, "VT_BLOB", p)

	p = PROPVARIANT{Value: PropVariantUnion{Caub: CAUB{CElems: 3, PElems: []uint8{1, 2, 3}}}}
	p.SetVt(VT_VECTOR_UI1)
	roundTrip(t, "VT_VECTOR|VT_UI1", p)

	p = PROPVARIANT{Value: PropVariantUnion{Calpwstr: CALPWSTR{CElems: 2, PElems: []*ndr.WSTR{wstr("s1"), wstr("s2")}}}}
	p.SetVt(VT_VECTOR_LPWSTR)
	roundTrip(t, "VT_VECTOR|VT_LPWSTR", p)
}

// TestPropVariantRecursiveArm exercises the VT_VECTOR|VT_VARIANT arm, whose elements are
// themselves PROPVARIANTs (a recursive type).
func TestPropVariantRecursiveArm(t *testing.T) {
	inner := PROPVARIANT{Value: PropVariantUnion{UlVal: 7}}
	inner.SetVt(VT_UI4)
	p := PROPVARIANT{Value: PropVariantUnion{Capropvar: CAPROPVARIANT{CElems: 1, PElems: []PROPVARIANT{inner}}}}
	p.SetVt(VT_VECTOR_VARIANT)
	roundTrip(t, "VT_VECTOR|VT_VARIANT", p)
}

// TestPropVariantEmpty exercises VT_EMPTY, which selects no union arm.
func TestPropVariantEmpty(t *testing.T) {
	var p PROPVARIANT
	p.SetVt(VT_EMPTY)
	roundTrip(t, "VT_EMPTY", p)
}

// TestMQMPSharedTypesRoundTrip covers the [MS-MQMQ] wire types reused by [MS-MQMP]:
// OBJECTID, XACTUOW, MULTICAST_ID, DL_ID, and each QUEUE_FORMAT union arm.
func TestMQMPSharedTypesRoundTrip(t *testing.T) {
	g := dtyp.GUID{Data1: 0x11223344, Data2: 0x5566, Data3: 0x7788, Data4: [8]byte{1, 2, 3, 4, 5, 6, 7, 8}}

	roundTrip(t, "OBJECTID", OBJECTID{Lineage: g, Uniquifier: 0xDEADBEEF})
	roundTrip(t, "XACTUOW", XACTUOW{Rgb: [16]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}})
	roundTrip(t, "MULTICAST_ID", MULTICAST_ID{MAddress: 0xC0A80001, MPort: 1801})
	roundTrip(t, "DL_ID", DL_ID{MDlGuid: g, MPwzDomain: wstr("example.com")})
}

// TestQueueFormatArms exercises each selectable QUEUE_FORMAT arm plus the UNKNOWN (empty)
// discriminant. SetQft keeps the outer and mirrored discriminants consistent.
func TestQueueFormatArms(t *testing.T) {
	g := dtyp.GUID{Data1: 0xAABBCCDD, Data2: 0x1122, Data3: 0x3344, Data4: [8]byte{9, 8, 7, 6, 5, 4, 3, 2}}

	var q QUEUE_FORMAT

	q = QUEUE_FORMAT{Value: QueueFormatUnion{MGPublicID: g}}
	roundTrip(t, "QF_PUBLIC", *q.SetQft(QUEUE_FORMAT_TYPE_PUBLIC))

	q = QUEUE_FORMAT{Value: QueueFormatUnion{MOPrivateID: OBJECTID{Lineage: g, Uniquifier: 42}}}
	roundTrip(t, "QF_PRIVATE", *q.SetQft(QUEUE_FORMAT_TYPE_PRIVATE))

	q = QUEUE_FORMAT{Value: QueueFormatUnion{MPDirectID: wstr(`OS:server\queue`)}}
	roundTrip(t, "QF_DIRECT", *q.SetQft(QUEUE_FORMAT_TYPE_DIRECT))

	q = QUEUE_FORMAT{Value: QueueFormatUnion{MGMachineID: g}}
	roundTrip(t, "QF_MACHINE", *q.SetQft(QUEUE_FORMAT_TYPE_MACHINE))

	q = QUEUE_FORMAT{Value: QueueFormatUnion{MDlID: DL_ID{MDlGuid: g, MPwzDomain: wstr("dom")}}}
	roundTrip(t, "QF_DL", *q.SetQft(QUEUE_FORMAT_TYPE_DL))

	q = QUEUE_FORMAT{Value: QueueFormatUnion{MMulticastID: MULTICAST_ID{MAddress: 0xE0000001, MPort: 2000}}}
	roundTrip(t, "QF_MULTICAST", *q.SetQft(QUEUE_FORMAT_TYPE_MULTICAST))

	q = QUEUE_FORMAT{}
	roundTrip(t, "QF_UNKNOWN", *q.SetQft(QUEUE_FORMAT_TYPE_UNKNOWN))
}
