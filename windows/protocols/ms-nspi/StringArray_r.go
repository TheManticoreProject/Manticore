package msnspi

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// StringArray_r ([MS-NSPI] 2.2.3.9) holds multiple 8-bit character strings. The IDL field
// is [size_is(cValues)] [string] char** lppszA — a [unique] pointer to a conformant array
// of [unique] pointers to ASCII strings.
type StringArray_r struct {
	CValues ndr.DWORD
	LppszA  []*ndr.STR `ndr:"unique,size_is=CValues,elem=unique"`
}
