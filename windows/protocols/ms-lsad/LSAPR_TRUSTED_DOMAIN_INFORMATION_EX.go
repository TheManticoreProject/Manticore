package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRUSTED_DOMAIN_INFORMATION_EX describes a trusted domain ([MS-LSAD] 2.2.7.9).
// Sid is a [unique] pointer to an RPC_SID.
type LSAPR_TRUSTED_DOMAIN_INFORMATION_EX struct {
	Name            dtyp.RPC_UNICODE_STRING
	FlatName        dtyp.RPC_UNICODE_STRING
	Sid             *dtyp.RPC_SID `ndr:"unique"`
	TrustDirection  ndr.DWORD
	TrustType       ndr.DWORD
	TrustAttributes ndr.DWORD
}
