package msnspi

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// StringsArray_r ([MS-NSPI] 2.2.5) is a counted array of 8-bit character strings. Strings
// is an embedded conformant array (maximum_count hoisted to the front of the structure)
// whose elements are [string] char* pointers; each element is a [unique] pointer to an
// ASCII string.
type StringsArray_r struct {
	Count   ndr.DWORD
	Strings []*ndr.STR `ndr:"conformant,size_is=Count,elem=unique"`
}
