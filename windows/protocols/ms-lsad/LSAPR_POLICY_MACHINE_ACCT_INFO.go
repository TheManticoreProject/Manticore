package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// LSAPR_POLICY_MACHINE_ACCT_INFO contains the machine account information of the server
// ([MS-LSAD] 2.2.4.17). Sid is a [unique] pointer to an RPC_SID.
type LSAPR_POLICY_MACHINE_ACCT_INFO struct {
	Rid ndr.DWORD
	Sid *msdtyp.RPC_SID `ndr:"unique"`
}
