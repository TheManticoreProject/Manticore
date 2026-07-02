package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_SID_INFORMATION wraps a single SID for a lookup request ([MS-LSAT] 2.2.17). Sid
// is a [unique] pointer to an RPC_SID.
type LSAPR_SID_INFORMATION struct {
	Sid *dtyp.RPC_SID `ndr:"unique"`
}
