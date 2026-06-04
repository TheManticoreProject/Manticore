package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_SID_INFORMATION holds a single SID pointer ([MS-SAMR] 2.2.7.4).
// SidPointer is a [unique] pointer to an RPC_SID.
type SAMPR_SID_INFORMATION struct {
	SidPointer *dtyp.RPC_SID `ndr:"unique"`
}
