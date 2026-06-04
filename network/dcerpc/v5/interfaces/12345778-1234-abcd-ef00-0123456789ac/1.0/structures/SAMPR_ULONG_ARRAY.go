package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_ULONG_ARRAY is a counted array of unsigned longs ([MS-SAMR] 2.2.3.8).
// Element is a [unique] pointer to a conformant array sized by Count.
type SAMPR_ULONG_ARRAY struct {
	Count   ndr.DWORD
	Element []ndr.DWORD `ndr:"unique,size_is=Count"`
}
