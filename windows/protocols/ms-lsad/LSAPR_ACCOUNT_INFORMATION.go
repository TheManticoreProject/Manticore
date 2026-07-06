package mslsad

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// LSAPR_ACCOUNT_INFORMATION identifies an account by SID ([MS-LSAD] 2.2.5.3). Sid is a
// [unique] pointer to an RPC_SID.
type LSAPR_ACCOUNT_INFORMATION struct {
	Sid *msdtyp.RPC_SID `ndr:"unique"`
}
