package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSA_FOREST_TRUST_COLLISION_RECORD describes a single forest-trust collision ([MS-LSAD]
// 2.2.7.26). Name is an LSA_UNICODE_STRING (the same wire form as RPC_UNICODE_STRING).
type LSA_FOREST_TRUST_COLLISION_RECORD struct {
	Index ndr.DWORD
	Type  LSA_FOREST_TRUST_COLLISION_RECORD_TYPE
	Flags ndr.DWORD
	Name  dtyp.RPC_UNICODE_STRING
}
