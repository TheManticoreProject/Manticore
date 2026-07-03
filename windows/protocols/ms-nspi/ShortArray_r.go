package msnspi

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ShortArray_r ([MS-NSPI] 2.2.3.3) holds multiple 16-bit integers. The IDL field is
// [size_is(cValues)] short int* lpi — a [unique] pointer to a conformant array of int16.
type ShortArray_r struct {
	CValues ndr.DWORD
	Lpi     []int16 `ndr:"unique,size_is=CValues"`
}
