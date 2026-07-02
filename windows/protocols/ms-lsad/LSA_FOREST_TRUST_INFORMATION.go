package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSA_FOREST_TRUST_INFORMATION carries the forest-trust records for a trusted domain
// ([MS-LSAD] 2.2.7.24). Entries is a [size_is(RecordCount)] array of [unique] pointers to
// LSA_FOREST_TRUST_RECORD.
type LSA_FOREST_TRUST_INFORMATION struct {
	RecordCount ndr.DWORD
	Entries     []*LSA_FOREST_TRUST_RECORD `ndr:"unique,size_is=RecordCount"`
}
