package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// LSAPR_TRUST_INFORMATION identifies a domain by name and SID ([MS-LSAD] 2.2.7.1).
// Sid is a [unique] pointer to an RPC_SID (pointer_default(unique)).
type LSAPR_TRUST_INFORMATION struct {
	Name dtyp.RPC_UNICODE_STRING
	Sid  *dtyp.RPC_SID `ndr:"unique"`
}
