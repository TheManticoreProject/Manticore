package mslsad

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// LSAPR_TRUST_INFORMATION identifies a domain by name and SID ([MS-LSAD] 2.2.7.1).
// Sid is a [unique] pointer to an RPC_SID (pointer_default(unique)).
type LSAPR_TRUST_INFORMATION struct {
	Name msdtyp.RPC_UNICODE_STRING
	Sid  *msdtyp.RPC_SID `ndr:"unique"`
}
