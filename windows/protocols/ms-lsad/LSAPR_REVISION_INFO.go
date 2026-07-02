package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_REVISION_INFO is the negotiation union exchanged by LsarOpenPolicy3 /
// LsarOpenPolicyWithCreds ([MS-LSAD] 2.2.2.1). Its switch_type is ULONG (a 4-byte
// discriminant, not a 16-bit enum), and version 1 selects the V1 arm. The discriminant is
// transmitted inline ahead of the selected arm ([C706] 14.3.8); set Tag to the InVersion /
// *OutVersion value before marshalling.
type LSAPR_REVISION_INFO struct {
	Tag ndr.DWORD              `ndr:"switch"`
	V1  LSAPR_REVISION_INFO_V1 `ndr:"case=1"`
}
