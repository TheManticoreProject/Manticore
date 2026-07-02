package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSA_FOREST_TRUST_INFORMATION2 carries the forest-trust records for a trusted domain
// ([MS-LSAD] 2.2.7.30), the v2 counterpart of LSA_FOREST_TRUST_INFORMATION. Entries is a
// [size_is(RecordCount)] array of [unique] pointers to LSA_FOREST_TRUST_RECORD2.
type LSA_FOREST_TRUST_INFORMATION2 struct {
	RecordCount ndr.DWORD
	Entries     []*LSA_FOREST_TRUST_RECORD2 `ndr:"unique,size_is=RecordCount"`
}
