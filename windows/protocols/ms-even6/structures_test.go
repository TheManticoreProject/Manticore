package mseven6

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals v (a pointer), unmarshals into out (a pointer to the same type), and
// leaves the comparison to the caller.
func roundTrip(t *testing.T, v any, out any) {
	t.Helper()
	raw, err := ndr.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%T): %v", v, err)
	}
	if err := ndr.Unmarshal(raw, out); err != nil {
		t.Fatalf("Unmarshal(%T): %v", v, err)
	}
}

func wstr(s string) *ndr.WSTR { w := ndr.WSTR(s); return &w }

var sampleGUID = dtyp.GUID{
	Data1: 0xf6beaff7,
	Data2: 0x1e19,
	Data3: 0x4fbb,
	Data4: [8]byte{0x9f, 0x8f, 0xb8, 0x9e, 0x20, 0x18, 0x33, 0x7c},
}

// TestRpcInfo_RoundTrip exercises RpcInfo ([MS-EVEN6] 2.2.1): three DWORDs, no pointers.
func TestRpcInfo_RoundTrip(t *testing.T) {
	in := RpcInfo{M_error: 5, M_subErr: 87, M_subErrParam: 2}
	var out RpcInfo
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("RpcInfo round-trip: got %+v want %+v", out, in)
	}
}

// TestEvtRpcQueryChannelInfo_RoundTrip exercises the [unique] string pointer Name plus the
// empty (nil) case ([MS-EVEN6] 2.2.11).
func TestEvtRpcQueryChannelInfo_RoundTrip(t *testing.T) {
	cases := []EvtRpcQueryChannelInfo{
		{Name: wstr(`Application`), Status: 0},
		{Name: nil, Status: 0x3A9F},
	}
	for _, in := range cases {
		var out EvtRpcQueryChannelInfo
		roundTrip(t, &in, &out)
		if !reflect.DeepEqual(in, out) {
			t.Fatalf("EvtRpcQueryChannelInfo round-trip: got %+v want %+v", out, in)
		}
	}
}

// TestScalarArrays_RoundTrip exercises the counted scalar-array structures, each a [unique]
// pointer to a conformant array bounded by Count ([MS-EVEN6] 2.2.7): BooleanArray,
// UInt32Array, UInt64Array — including the empty (nil) case.
func TestScalarArrays_RoundTrip(t *testing.T) {
	ba := BooleanArray{Count: 3, Ptr: []bool{true, false, true}}
	var baOut BooleanArray
	roundTrip(t, &ba, &baOut)
	if !reflect.DeepEqual(ba, baOut) {
		t.Fatalf("BooleanArray round-trip: got %+v want %+v", baOut, ba)
	}

	u32 := UInt32Array{Count: 2, Ptr: []ndr.DWORD{0xdeadbeef, 0x00c0ffee}}
	var u32Out UInt32Array
	roundTrip(t, &u32, &u32Out)
	if !reflect.DeepEqual(u32, u32Out) {
		t.Fatalf("UInt32Array round-trip: got %+v want %+v", u32Out, u32)
	}

	u64 := UInt64Array{Count: 2, Ptr: []uint64{0x1122334455667788, 0}}
	var u64Out UInt64Array
	roundTrip(t, &u64, &u64Out)
	if !reflect.DeepEqual(u64, u64Out) {
		t.Fatalf("UInt64Array round-trip: got %+v want %+v", u64Out, u64)
	}

	empty := UInt32Array{Count: 0, Ptr: nil}
	var emptyOut UInt32Array
	roundTrip(t, &empty, &emptyOut)
	if !reflect.DeepEqual(empty, emptyOut) {
		t.Fatalf("empty UInt32Array round-trip: got %+v want %+v", emptyOut, empty)
	}
}

// TestStringArray_RoundTrip exercises StringArray ([MS-EVEN6] 2.2.8): a [unique] pointer to
// a conformant array of [unique] wide-string pointers.
func TestStringArray_RoundTrip(t *testing.T) {
	in := StringArray{Count: 2, Ptr: []*ndr.WSTR{wstr(`Security`), wstr(`System`)}}
	var out StringArray
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("StringArray round-trip: got %+v want %+v", out, in)
	}
}

// TestGuidArray_RoundTrip exercises GuidArray ([MS-EVEN6] 2.2.9): a [unique] pointer to a
// conformant array of 16-octet GUIDs.
func TestGuidArray_RoundTrip(t *testing.T) {
	in := GuidArray{Count: 1, Ptr: []dtyp.GUID{sampleGUID}}
	var out GuidArray
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("GuidArray round-trip: got %+v want %+v", out, in)
	}
}

// TestEvtRpcVariant_RoundTrip exercises the switch_is(type) union ([MS-EVEN6] 2.2.6) across
// scalar, pointer, and array arms. Type and Field.Tag are set together (this codec transmits
// the discriminant inline as well); only the selected arm is populated so DeepEqual holds.
func TestEvtRpcVariant_RoundTrip(t *testing.T) {
	cases := []EvtRpcVariant{
		{Type: EvtRpcVarTypeNull, Field: EvtRpcVariant_Field{Tag: ndr.DWORD(EvtRpcVarTypeNull)}},
		{Type: EvtRpcVarTypeBoolean, Field: EvtRpcVariant_Field{Tag: ndr.DWORD(EvtRpcVarTypeBoolean), BooleanVal: true}},
		{Type: EvtRpcVarTypeUInt32, Field: EvtRpcVariant_Field{Tag: ndr.DWORD(EvtRpcVarTypeUInt32), Uint32Val: 0x41424344}},
		{Type: EvtRpcVarTypeUInt64, Field: EvtRpcVariant_Field{Tag: ndr.DWORD(EvtRpcVarTypeUInt64), Uint64Val: 0x1122334455667788}},
		{Type: EvtRpcVarTypeString, Field: EvtRpcVariant_Field{Tag: ndr.DWORD(EvtRpcVarTypeString), StringVal: wstr(`Application`)}},
		{Type: EvtRpcVarTypeGuid, Field: EvtRpcVariant_Field{Tag: ndr.DWORD(EvtRpcVarTypeGuid), GuidVal: &sampleGUID}},
		{Type: EvtRpcVarTypeUInt32Array, Field: EvtRpcVariant_Field{Tag: ndr.DWORD(EvtRpcVarTypeUInt32Array), Uint32Array: UInt32Array{Count: 2, Ptr: []ndr.DWORD{1, 2}}}},
	}
	for _, in := range cases {
		var out EvtRpcVariant
		roundTrip(t, &in, &out)
		if !reflect.DeepEqual(in, out) {
			t.Fatalf("EvtRpcVariant(type=%d) round-trip: got %+v want %+v", in.Type, out, in)
		}
	}
}

// TestEvtRpcVariantList_RoundTrip exercises the [unique] pointer to a conformant array of
// unions ([MS-EVEN6] 2.2.10).
func TestEvtRpcVariantList_RoundTrip(t *testing.T) {
	in := EvtRpcVariantList{
		Count: 2,
		Props: []EvtRpcVariant{
			{Type: EvtRpcVarTypeUInt32, Field: EvtRpcVariant_Field{Tag: ndr.DWORD(EvtRpcVarTypeUInt32), Uint32Val: 7}},
			{Type: EvtRpcVarTypeString, Field: EvtRpcVariant_Field{Tag: ndr.DWORD(EvtRpcVarTypeString), StringVal: wstr(`x`)}},
		},
	}
	var out EvtRpcVariantList
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("EvtRpcVariantList round-trip: got %+v want %+v", out, in)
	}
}
