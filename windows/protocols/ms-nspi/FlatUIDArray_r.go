package msnspi

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// FlatUIDArray_r ([MS-NSPI] 2.2.3.7) holds multiple GUIDs. The IDL field is
// [size_is(cValues)] FlatUID_r** lpguid — a [unique] pointer to a conformant array of
// [unique] pointers to FlatUID_r.
type FlatUIDArray_r struct {
	CValues ndr.DWORD
	Lpguid  []*FlatUID_r `ndr:"unique,size_is=CValues,elem=unique"`
}
