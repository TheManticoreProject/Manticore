package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSAPR_REVISION_INFO_V1 is version 1 of the LSAPR_REVISION_INFO negotiation union
// ([MS-LSAD] 2.2.2.2): the client/server revision and a bitmask of supported features.
type LSAPR_REVISION_INFO_V1 struct {
	Revision          ndr.DWORD
	SupportedFeatures ndr.DWORD
}
