package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_REVISION_INFO_V1 carries the client/server revision and supported
// feature flags ([MS-SAMR] 2.2.7.13).
type SAMPR_REVISION_INFO_V1 struct {
	Revision          ndr.DWORD
	SupportedFeatures ndr.DWORD
}
