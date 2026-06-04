package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SAMPR_REVISION_INFO is a switch_type(unsigned long) union with a single
// defined arm ([MS-SAMR] 2.2.7.14). Tag is the discriminant carried inline.
type SAMPR_REVISION_INFO struct {
	Tag ndr.DWORD              `ndr:"switch"`
	V1  SAMPR_REVISION_INFO_V1 `ndr:"case=1"`
}
