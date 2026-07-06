package mslsad

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// LSAPR_POLICY_ACCOUNT_DOM_INFO contains information about the account domain ([MS-LSAD]
// 2.2.4.6). DomainSid is a [unique] pointer to an RPC_SID.
type LSAPR_POLICY_ACCOUNT_DOM_INFO struct {
	DomainName msdtyp.RPC_UNICODE_STRING
	DomainSid  *msdtyp.RPC_SID `ndr:"unique"`
}
