package msrprn

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// RPC_PrintPropertyValue ([MS-RPRN] 2.2.1.13.2). The ePropertyType discriminant selects
// the arm of the value union; per [C706] 14.3.8 the discriminant is transmitted both as
// this struct field and inline in the union.
type RPC_PrintPropertyValue struct {
	EPropertyType RPC_EPrintPropertyType
	Value         RPC_PrintPropertyValue_Value
}

// RPC_PrintPropertyValue_Value is the [switch_is(ePropertyType)] union of RPC_PrintPropertyValue.
// The switch discriminant is the RPC_EPrintPropertyType enum (16-bit under NDR20, [C706] 14.3.6).
type RPC_PrintPropertyValue_Value struct {
	Tag            RPC_EPrintPropertyType `ndr:"switch,enum"`
	PropertyString *ndr.WSTR              `ndr:"case=1,unique"`
	PropertyInt32  int32                  `ndr:"case=2"`
	PropertyInt64  int64                  `ndr:"case=3"`
	PropertyByte   uint8                  `ndr:"case=4"`
	PropertyBlob   RPC_PrintPropertyBlob  `ndr:"case=5"`
}

// RPC_PrintPropertyBlob is the anonymous propertyBlob arm ([MS-RPRN] 2.2.1.13.2):
// a counted byte buffer { DWORD cbBuf; [size_is(cbBuf)] BYTE *pBuf; }.
type RPC_PrintPropertyBlob struct {
	CbBuf ndr.DWORD
	PBuf  []uint8 `ndr:"unique,size_is=CbBuf"`
}
