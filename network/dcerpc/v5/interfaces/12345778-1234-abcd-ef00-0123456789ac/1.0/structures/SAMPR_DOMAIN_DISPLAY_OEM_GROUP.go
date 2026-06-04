package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_DOMAIN_DISPLAY_OEM_GROUP holds a single OEM (ASCII) group entry returned by a
// display query ([MS-SAMR] 2.2.8.6). OemAccountName is an RPC_STRING defined by the base
// family.
type SAMPR_DOMAIN_DISPLAY_OEM_GROUP struct {
	Index          ndr.DWORD
	OemAccountName RPC_STRING
}
