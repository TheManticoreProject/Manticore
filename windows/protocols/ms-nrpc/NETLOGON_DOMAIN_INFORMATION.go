package msnrpc

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// NETLOGON_DOMAIN_INFORMATION is a discriminated union ([MS-NRPC]); the discriminant precedes the
// selected arm ([C706] 14.3.8). Both arms are P* pointers in the IDL, so they are modeled
// as [unique] pointers: a referent id followed by the deferred body.
type NETLOGON_DOMAIN_INFORMATION struct {
	Tag           ndr.DWORD                 `ndr:"switch"`
	DomainInfo    *NETLOGON_DOMAIN_INFO     `ndr:"case=1,unique"`
	LsaPolicyInfo *NETLOGON_LSA_POLICY_INFO `ndr:"case=2,unique"`
}
