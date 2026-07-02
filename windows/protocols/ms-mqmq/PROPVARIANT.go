package msmqmq

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// PROPVARIANT is the Message Queuing property variant ([MS-MQMQ] 2.2.12): a VARTYPE
// discriminant (vt), three reserved fields, and a [switch_is(vt)] union carrying the
// typed value. It is the property value used throughout [MS-MQDS] (the aProp/apVar
// property arrays and MQPROPERTYRESTRICTION).
//
// Wire modelling. The IDL union is non-encapsulated (its discriminant is the enclosing
// struct's vt field). The codec's declarative unions carry their own discriminant, so the
// union is modelled as the Value field whose own Vt switch mirrors the outer Vt. Keep the
// two equal (SetVt does this); a mismatch would marshal an inconsistent discriminant.
type PROPVARIANT struct {
	Vt         VARTYPE
	WReserved1 uint8
	WReserved2 uint8
	WReserved3 uint32
	Value      PropVariantUnion
}

// SetVt sets both the outer discriminant and the union's mirrored discriminant so they
// stay consistent, then returns the receiver for chaining.
func (p *PROPVARIANT) SetVt(vt VARTYPE) *PROPVARIANT {
	p.Vt = vt
	p.Value.Vt = vt
	return p
}

// PropVariantUnion is the [switch_is(vt)] union of a PROPVARIANT ([MS-MQMQ] 2.2.12). Each
// arm's case value is the VARTYPE (or VT_VECTOR|<base>) that selects it. VT_EMPTY and
// VT_NULL select no arm (an empty union body). Pointer arms use the unique referent form
// mandated by the interface's pointer_default(unique).
type PropVariantUnion struct {
	Vt VARTYPE `ndr:"switch"`

	CVal    int8   `ndr:"case=16"` // VT_I1
	BVal    uint8  `ndr:"case=17"` // VT_UI1
	IVal    int16  `ndr:"case=2"`  // VT_I2
	UiVal   uint16 `ndr:"case=18"` // VT_UI2
	LVal    int32  `ndr:"case=3"`  // VT_I4
	UlVal   uint32 `ndr:"case=19"` // VT_UI4
	HVal    int64  `ndr:"case=20"` // VT_I8  (LARGE_INTEGER)
	UhVal   uint64 `ndr:"case=21"` // VT_UI8 (ULARGE_INTEGER)
	BoolVal int16  `ndr:"case=11"` // VT_BOOL (VARIANT_BOOL)

	Puuid   *dtyp.GUID `ndr:"case=72,unique"` // VT_CLSID
	Blob    BLOB       `ndr:"case=65"`        // VT_BLOB
	PwszVal *ndr.WSTR  `ndr:"case=31,unique"` // VT_LPWSTR

	Caub      CAUB          `ndr:"case=0x1011"` // VT_VECTOR|VT_UI1
	Caui      CAUI          `ndr:"case=0x1012"` // VT_VECTOR|VT_UI2
	Cal       CAL           `ndr:"case=0x1003"` // VT_VECTOR|VT_I4
	Caul      CAUL          `ndr:"case=0x1013"` // VT_VECTOR|VT_UI4
	Cauh      CAUH          `ndr:"case=0x1015"` // VT_VECTOR|VT_UI8
	Cauuid    CACLSID       `ndr:"case=0x1048"` // VT_VECTOR|VT_CLSID
	Calpwstr  CALPWSTR      `ndr:"case=0x101F"` // VT_VECTOR|VT_LPWSTR
	Capropvar CAPROPVARIANT `ndr:"case=0x100C"` // VT_VECTOR|VT_VARIANT
}
