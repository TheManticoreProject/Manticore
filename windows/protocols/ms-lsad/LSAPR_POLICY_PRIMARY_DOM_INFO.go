package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_POLICY_PRIMARY_DOM_INFO contains information about the primary domain ([MS-LSAD]
// 2.2.4.5). Sid is a [unique] pointer to an RPC_SID.
type LSAPR_POLICY_PRIMARY_DOM_INFO struct {
	Name dtyp.RPC_UNICODE_STRING
	Sid  *dtyp.RPC_SID `ndr:"unique"`
}
