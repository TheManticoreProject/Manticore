package msnrpc

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// NETLOGON_CONTROL_QUERY_INFORMATION is a discriminated union ([MS-NRPC]); the discriminant precedes the
// selected arm ([C706] 14.3.8). Every arm is a PNETLOGON_INFO_<n> pointer in the IDL, so
// it is modeled as a [unique] pointer: a referent id followed by its deferred body.
type NETLOGON_CONTROL_QUERY_INFORMATION struct {
	Tag           ndr.DWORD        `ndr:"switch"`
	NetlogonInfo1 *NETLOGON_INFO_1 `ndr:"case=1,unique"`
	NetlogonInfo2 *NETLOGON_INFO_2 `ndr:"case=2,unique"`
	NetlogonInfo3 *NETLOGON_INFO_3 `ndr:"case=3,unique"`
	NetlogonInfo4 *NETLOGON_INFO_4 `ndr:"case=4,unique"`
}
