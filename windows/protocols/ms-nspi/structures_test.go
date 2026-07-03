package msnspi

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts the
// result is deeply equal to in. This is the wire-shape acceptance gate for the MS-NSPI
// NDR structures in the absence of a live NSPI server (Go round-trip is necessary but not
// sufficient for wire correctness — see the dcerpc-interface-structure skill).
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

func strp(s string) *ndr.STR   { v := ndr.STR(s); return &v }
func wstrp(s string) *ndr.WSTR { v := ndr.WSTR(s); return &v }

func testFlatUID(b byte) FlatUID_r {
	var u FlatUID_r
	for i := range u.Ab {
		u.Ab[i] = b
	}
	return u
}

// TestScalarStructs covers the plain fixed-layout structures (no pointers or arrays).
func TestScalarStructs(t *testing.T) {
	roundTrip(t, "FlatUID_r", testFlatUID(0xAB))
	roundTrip(t, "STAT", STAT{
		SortType: 0, ContainerID: 1, CurrentRec: 2, Delta: -3, NumPos: 4,
		TotalRecs: 5, CodePage: 1252, TemplateLocale: 0x409, SortLocale: 0x409,
	})
	roundTrip(t, "BitMaskRestriction_r", BitMaskRestriction_r{RelBMR: 0, UlPropTag: 0x30140003, UlMask: 0x40})
	roundTrip(t, "ComparePropsRestriction_r", ComparePropsRestriction_r{Relop: 4, UlPropTag1: 0x0001, UlPropTag2: 0x0002})
	roundTrip(t, "SizeRestriction_r", SizeRestriction_r{Relop: 2, UlPropTag: 0x0C1F001E, Cb: 16})
	roundTrip(t, "ExistRestriction_r", ExistRestriction_r{UlReserved1: 0, UlPropTag: 0x3001001F, UlReserved2: 0})
}

// TestScalarArrays covers the single-pointer conformant arrays of scalars/structs.
func TestScalarArrays(t *testing.T) {
	roundTrip(t, "Binary_r", Binary_r{Cb: 4, Lpb: []uint8{0xDE, 0xAD, 0xBE, 0xEF}})
	roundTrip(t, "Binary_r/nil", Binary_r{Cb: 0, Lpb: nil})
	roundTrip(t, "ShortArray_r", ShortArray_r{CValues: 3, Lpi: []int16{-1, 0, 32000}})
	roundTrip(t, "LongArray_r", LongArray_r{CValues: 2, Lpl: []int32{-100, 100}})
	roundTrip(t, "DateTimeArray_r", DateTimeArray_r{CValues: 1, Lpft: []FILETIME{{DwLowDateTime: 1, DwHighDateTime: 2}}})
	roundTrip(t, "PropertyName_r", PropertyName_r{Lpguid: &FlatUID_r{Ab: [16]byte{1, 2, 3}}, UlReserved: 0, LID: 0x8004})
	roundTrip(t, "BinaryArray_r", BinaryArray_r{CValues: 2, Lpbin: []Binary_r{
		{Cb: 2, Lpb: []uint8{0x01, 0x02}},
		{Cb: 0, Lpb: nil},
	}})
}

// TestPointerElementArrays covers arrays of pointers ([]*T with elem=unique): the
// double-pointer sites collapsed by idlgen and reconciled by hand.
func TestPointerElementArrays(t *testing.T) {
	roundTrip(t, "FlatUIDArray_r", FlatUIDArray_r{CValues: 2, Lpguid: []*FlatUID_r{
		{Ab: [16]byte{0xAA}}, {Ab: [16]byte{0xBB}},
	}})
	roundTrip(t, "StringArray_r", StringArray_r{CValues: 2, LppszA: []*ndr.STR{strp("alpha"), strp("beta")}})
	roundTrip(t, "WStringArray_r", WStringArray_r{CValues: 2, LppszW: []*ndr.WSTR{wstrp("uno"), wstrp("dos")}})
	roundTrip(t, "StringsArray_r", StringsArray_r{Count: 2, Strings: []*ndr.STR{strp("/o=x"), strp("/o=y")}})
	roundTrip(t, "WStringsArray_r", WStringsArray_r{Count: 1, Strings: []*ndr.WSTR{wstrp("wide")}})
}

// TestPropValUnion covers several arms of the non-encapsulated PROP_VAL_UNION; only the
// arm selected by Tag is transmitted and must survive the round trip.
func TestPropValUnion(t *testing.T) {
	roundTrip(t, "PV/int16", PROP_VAL_UNION{Tag: 0x0002, I: -7})
	roundTrip(t, "PV/int32", PROP_VAL_UNION{Tag: 0x0003, L: 0x11223344})
	roundTrip(t, "PV/bool", PROP_VAL_UNION{Tag: 0x000B, B: 1})
	roundTrip(t, "PV/str8", PROP_VAL_UNION{Tag: 0x001E, LpszA: strp("smith")})
	roundTrip(t, "PV/strW", PROP_VAL_UNION{Tag: 0x001F, LpszW: wstrp("smith")})
	roundTrip(t, "PV/bin", PROP_VAL_UNION{Tag: 0x0102, Bin: Binary_r{Cb: 3, Lpb: []uint8{1, 2, 3}}})
	roundTrip(t, "PV/guid", PROP_VAL_UNION{Tag: 0x0048, Lpguid: &FlatUID_r{Ab: [16]byte{9}}})
	roundTrip(t, "PV/time", PROP_VAL_UNION{Tag: 0x0040, Ft: FILETIME{DwLowDateTime: 7, DwHighDateTime: 8}})
	roundTrip(t, "PV/err", PROP_VAL_UNION{Tag: 0x000A, Err: 0x0004010F})
	roundTrip(t, "PV/mvint32", PROP_VAL_UNION{Tag: 0x1003, MVl: LongArray_r{CValues: 2, Lpl: []int32{1, 2}}})
	roundTrip(t, "PV/mvstr8", PROP_VAL_UNION{Tag: 0x101E, MVszA: StringArray_r{CValues: 1, LppszA: []*ndr.STR{strp("x")}}})
}

// TestPropertyRowAndSet covers PropertyValue_r → PropertyRow_r → PropertyRowSet_r nesting,
// exercising the embedded union inside a conformant array of structs.
func TestPropertyRowAndSet(t *testing.T) {
	row := PropertyRow_r{
		Reserved: 0,
		CValues:  2,
		LpProps: []PropertyValue_r{
			{UlPropTag: 0x3001001F, UlReserved: 0, Value: PROP_VAL_UNION{Tag: 0x001F, LpszW: wstrp("Display Name")}},
			{UlPropTag: 0x39000003, UlReserved: 0, Value: PROP_VAL_UNION{Tag: 0x0003, L: 6}},
		},
	}
	roundTrip(t, "PropertyRow_r", row)
	roundTrip(t, "PropertyRowSet_r", PropertyRowSet_r{CRows: 1, ARow: []PropertyRow_r{row}})
	roundTrip(t, "PropertyNameSet_r", PropertyNameSet_r{CNames: 2, ANames: []PropertyName_r{
		{Lpguid: &FlatUID_r{Ab: [16]byte{1}}, UlReserved: 0, LID: 0x8001},
		{Lpguid: &FlatUID_r{Ab: [16]byte{2}}, UlReserved: 0, LID: 0x8002},
	}})
}

// TestRestriction covers the recursive Restriction_r / RestrictionUnion_r: an AND node
// whose sub-restrictions are an Exist leaf and a Property leaf carrying a PropertyValue_r.
func TestRestriction(t *testing.T) {
	exist := Restriction_r{Rt: 0x00000008, Res: RestrictionUnion_r{Tag: 0x00000008,
		ResExist: ExistRestriction_r{UlPropTag: 0x3001001F}}}
	prop := Restriction_r{Rt: 0x00000004, Res: RestrictionUnion_r{Tag: 0x00000004,
		ResProperty: PropertyRestriction_r{Relop: 4, UlPropTag: 0x3001001F,
			LpProp: &PropertyValue_r{UlPropTag: 0x3001001F, Value: PROP_VAL_UNION{Tag: 0x001F, LpszW: wstrp("target")}}}}}
	and := Restriction_r{Rt: 0x00000000, Res: RestrictionUnion_r{Tag: 0x00000000,
		ResAnd: AndRestriction_r{CRes: 2, LpRes: []Restriction_r{exist, prop}}}}
	roundTrip(t, "Restriction_r/exist", exist)
	roundTrip(t, "Restriction_r/property", prop)
	roundTrip(t, "Restriction_r/and", and)
	roundTrip(t, "Restriction_r/not", Restriction_r{Rt: 0x00000002, Res: RestrictionUnion_r{Tag: 0x00000002,
		ResNot: NotRestriction_r{LpRes: &exist}}})
}

// TestFILETIMEAlias guards the dtyp alias so a wrong redefinition is caught at compile+run.
func TestFILETIMEAlias(t *testing.T) {
	var ft FILETIME = dtyp.FILETIME{DwLowDateTime: 0x11111111, DwHighDateTime: 0x22222222}
	roundTrip(t, "FILETIME", ft)
}

// ptagPtrWrap embeds PropertyTagArray_r behind a [unique] pointer, the way real
// request/response structs reference it, so the conformant-varying + referent handling is
// exercised end to end.
type ptagPtrWrap struct {
	P *PropertyTagArray_r `ndr:"unique"`
}

// TestPropertyTagArray round-trips the counted property-tag array both directly and behind a
// [unique] pointer, including the NULL and empty cases.
func TestPropertyTagArray(t *testing.T) {
	roundTrip(t, "PropertyTagArray_r", PropertyTagArray_r{CValues: 3, AulPropTag: []ndr.DWORD{0x3001001F, 0x39000003, 0x0FFF0102}})
	roundTrip(t, "ptagPtrWrap", ptagPtrWrap{P: &PropertyTagArray_r{CValues: 2, AulPropTag: []ndr.DWORD{0x0A, 0x0B}}})
	roundTrip(t, "ptagPtrWrap/nil", ptagPtrWrap{P: nil})
}

// TestPropValUnionReservedDLabel covers the second discriminant label of the reserved arm
// (0x0000000D), which shares a long with 0x00000001.
func TestPropValUnionReservedDLabel(t *testing.T) {
	roundTrip(t, "PV/reserved0D", PROP_VAL_UNION{Tag: 0x000D, LReserved0D: 0x0BADF00D})
	roundTrip(t, "PV/reserved01", PROP_VAL_UNION{Tag: 0x0001, LReserved: 0x0000CAFE})
}

// TestSetDiscriminants verifies the discriminant helpers fill in the union tags a caller
// would otherwise leave zero, so the selected arm survives a marshal/unmarshal round trip.
func TestSetDiscriminants(t *testing.T) {
	// PropertyValue_r: caller sets UlPropTag + arm but not Value.Tag.
	pv := PropertyValue_r{UlPropTag: 0x3001001F, Value: PROP_VAL_UNION{LpszW: wstrp("name")}}
	pv.SetDiscriminant()
	if pv.Value.Tag != 0x001F {
		t.Fatalf("SetDiscriminant: Tag = 0x%x, want 0x1F", pv.Value.Tag)
	}
	roundTrip(t, "PropertyValue_r/normalized", pv)

	// Restriction_r: an AND of (Property, Exist) with only rt + arms set, no Tag anywhere.
	r := Restriction_r{Rt: RestrictionTypeAnd, Res: RestrictionUnion_r{ResAnd: AndRestriction_r{
		CRes: 2, LpRes: []Restriction_r{
			{Rt: RestrictionTypeProperty, Res: RestrictionUnion_r{ResProperty: PropertyRestriction_r{
				Relop: 4, UlPropTag: 0x3001001F,
				LpProp: &PropertyValue_r{UlPropTag: 0x3001001F, Value: PROP_VAL_UNION{LpszW: wstrp("x")}}}}},
			{Rt: RestrictionTypeExist, Res: RestrictionUnion_r{ResExist: ExistRestriction_r{UlPropTag: 0x39000003}}},
		}}}}
	r.SetDiscriminants()
	if r.Res.Tag != int32(RestrictionTypeAnd) ||
		r.Res.ResAnd.LpRes[0].Res.Tag != int32(RestrictionTypeProperty) ||
		r.Res.ResAnd.LpRes[0].Res.ResProperty.LpProp.Value.Tag != 0x001F ||
		r.Res.ResAnd.LpRes[1].Res.Tag != int32(RestrictionTypeExist) {
		t.Fatalf("SetDiscriminants did not populate nested tags: %+v", r)
	}
	roundTrip(t, "Restriction_r/normalized", r)

	// PropertyRow_r: per-element discriminants.
	row := PropertyRow_r{CValues: 1, LpProps: []PropertyValue_r{
		{UlPropTag: 0x39000003, Value: PROP_VAL_UNION{L: 6}},
	}}
	row.SetDiscriminants()
	if row.LpProps[0].Value.Tag != 0x0003 {
		t.Fatalf("PropertyRow_r.SetDiscriminants: Tag = 0x%x, want 0x3", row.LpProps[0].Value.Tag)
	}
}
