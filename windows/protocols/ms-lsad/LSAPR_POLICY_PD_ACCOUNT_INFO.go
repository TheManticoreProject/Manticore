package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_POLICY_PD_ACCOUNT_INFO is an unused policy information class ([MS-LSAD]
// 2.2.4.7).
type LSAPR_POLICY_PD_ACCOUNT_INFO struct {
	Name dtyp.RPC_UNICODE_STRING
}
