package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_ACCOUNT_INFORMATION identifies an account by SID ([MS-LSAD] 2.2.5.3). Sid is a
// [unique] pointer to an RPC_SID.
type LSAPR_ACCOUNT_INFORMATION struct {
	Sid *dtyp.RPC_SID `ndr:"unique"`
}
