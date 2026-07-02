package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_POLICY_ACCOUNT_DOM_INFO contains information about the account domain ([MS-LSAD]
// 2.2.4.6). DomainSid is a [unique] pointer to an RPC_SID.
type LSAPR_POLICY_ACCOUNT_DOM_INFO struct {
	DomainName dtyp.RPC_UNICODE_STRING
	DomainSid  *dtyp.RPC_SID `ndr:"unique"`
}
