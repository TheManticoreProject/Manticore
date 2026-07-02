package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSA_FOREST_TRUST_COLLISION_INFORMATION carries the set of collisions detected when
// setting forest-trust information ([MS-LSAD] 2.2.7.27). Entries is a
// [size_is(RecordCount)] array of [unique] pointers to LSA_FOREST_TRUST_COLLISION_RECORD.
type LSA_FOREST_TRUST_COLLISION_INFORMATION struct {
	RecordCount ndr.DWORD
	Entries     []*LSA_FOREST_TRUST_COLLISION_RECORD `ndr:"unique,size_is=RecordCount"`
}
