package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_TRUSTED_DOMAIN_INFORMATION_EX2 extends LSAPR_TRUSTED_DOMAIN_INFORMATION_EX with
// forest-trust information ([MS-LSAD] 2.2.7.10). Sid is a [unique] pointer to an
// RPC_SID; ForestTrustInfo is a [unique] pointer to a conformant byte array sized by
// ForestTrustLength.
type LSAPR_TRUSTED_DOMAIN_INFORMATION_EX2 struct {
	Name              dtyp.RPC_UNICODE_STRING
	FlatName          dtyp.RPC_UNICODE_STRING
	Sid               *dtyp.RPC_SID `ndr:"unique"`
	TrustDirection    ndr.DWORD
	TrustType         ndr.DWORD
	TrustAttributes   ndr.DWORD
	ForestTrustLength ndr.DWORD
	ForestTrustInfo   []byte `ndr:"unique,size_is=ForestTrustLength"`
}
