package mseven6

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// EvtRpcVariant ([MS-EVEN6] 2.2.6) carries a tagged value. Type is the discriminant (a
// [v1_enum], 32-bit) and Flags a bitmask; Field is the switch_is(type) union. Set both Type
// and Field.Tag to the same EvtRpcVariantType before marshalling — this codec transmits the
// union discriminant inline as well (see network/dcerpc/ndr/union.go).
type EvtRpcVariant struct {
	Type  EvtRpcVariantType
	Flags ndr.DWORD
	Field EvtRpcVariant_Field
}

// EvtRpcVariant_Field is the switch_is(type) union of EvtRpcVariant ([MS-EVEN6] 2.2.6). The
// discriminant Tag precedes the selected arm ([C706] 14.3.8) and is a 32-bit value matching
// the v1_enum EvtRpcVariantType. The stringVal and guidVal arms are [unique] pointers under
// pointer_default(unique); the *Array arms are inline counted-array structures.
type EvtRpcVariant_Field struct {
	Tag          ndr.DWORD    `ndr:"switch"`
	NullVal      int32        `ndr:"case=0"`
	BooleanVal   bool         `ndr:"case=1"`
	Uint32Val    ndr.DWORD    `ndr:"case=2"`
	Uint64Val    uint64       `ndr:"case=3"`
	StringVal    *ndr.WSTR    `ndr:"case=4,unique"`
	GuidVal      *msdtyp.GUID `ndr:"case=5,unique"`
	BooleanArray BooleanArray `ndr:"case=6"`
	Uint32Array  UInt32Array  `ndr:"case=7"`
	Uint64Array  UInt64Array  `ndr:"case=8"`
	StringArray  StringArray  `ndr:"case=9"`
	GuidArray    GuidArray    `ndr:"case=10"`
}
