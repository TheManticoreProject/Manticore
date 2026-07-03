package msnspi

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// WStringsArray_r ([MS-NSPI] 2.2.7) is a counted array of Unicode strings. Strings is an
// embedded conformant array (maximum_count hoisted to the front of the structure) whose
// elements are [string] wchar_t* pointers; each element is a [unique] pointer to a Unicode
// string.
type WStringsArray_r struct {
	Count   ndr.DWORD
	Strings []*ndr.WSTR `ndr:"conformant,size_is=Count,elem=unique"`
}
