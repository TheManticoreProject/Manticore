package msnrpc

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// NETLOGON_WORKSTATION_INFORMATION is a discriminated union ([MS-NRPC]); the discriminant precedes the
// selected arm ([C706] 14.3.8). Both arms are PNETLOGON_WORKSTATION_INFO pointers in the
// IDL, so they are modeled as [unique] pointers: a referent id followed by the deferred body.
type NETLOGON_WORKSTATION_INFORMATION struct {
	Tag             ndr.DWORD                  `ndr:"switch"`
	WorkstationInfo *NETLOGON_WORKSTATION_INFO `ndr:"case=1,unique"`
	LsaPolicyInfo   *NETLOGON_WORKSTATION_INFO `ndr:"case=2,unique"`
}
