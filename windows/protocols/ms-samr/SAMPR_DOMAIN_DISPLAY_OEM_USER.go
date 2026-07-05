package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_DOMAIN_DISPLAY_OEM_USER holds a single OEM (ASCII) user entry returned by a
// display query ([MS-SAMR] 2.2.8.5). OemAccountName is an RPC_STRING defined by the base
// family.
type SAMPR_DOMAIN_DISPLAY_OEM_USER struct {
	Index          ndr.DWORD
	OemAccountName RPC_STRING
}
