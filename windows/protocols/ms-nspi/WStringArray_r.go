package msnspi

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// WStringArray_r ([MS-NSPI] 2.2.3.11) holds multiple Unicode strings. The IDL field is
// [size_is(cValues)] [string] wchar_t** lppszW — a [unique] pointer to a conformant array
// of [unique] pointers to Unicode strings.
type WStringArray_r struct {
	CValues ndr.DWORD
	LppszW  []*ndr.WSTR `ndr:"unique,size_is=CValues,elem=unique"`
}
