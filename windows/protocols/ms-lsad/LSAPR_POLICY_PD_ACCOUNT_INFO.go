package mslsad

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// LSAPR_POLICY_PD_ACCOUNT_INFO is an unused policy information class ([MS-LSAD]
// 2.2.4.7).
type LSAPR_POLICY_PD_ACCOUNT_INFO struct {
	Name msdtyp.RPC_UNICODE_STRING
}
