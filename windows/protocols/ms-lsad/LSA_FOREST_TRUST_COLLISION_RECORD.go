package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// LSA_FOREST_TRUST_COLLISION_RECORD describes a single forest-trust collision ([MS-LSAD]
// 2.2.7.26). Name is an LSA_UNICODE_STRING (the same wire form as RPC_UNICODE_STRING).
type LSA_FOREST_TRUST_COLLISION_RECORD struct {
	Index ndr.DWORD
	Type  LSA_FOREST_TRUST_COLLISION_RECORD_TYPE `ndr:"enum"`
	Flags ndr.DWORD
	Name  msdtyp.RPC_UNICODE_STRING
}
