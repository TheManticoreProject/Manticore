package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSA_FOREST_TRUST_BINARY_DATA holds an opaque forest-trust record payload ([MS-LSAD]
// 2.2.7.22). Buffer is a [unique] pointer to a conformant byte array sized by Length.
type LSA_FOREST_TRUST_BINARY_DATA struct {
	Length ndr.DWORD
	Buffer []byte `ndr:"unique,size_is=Length"`
}
