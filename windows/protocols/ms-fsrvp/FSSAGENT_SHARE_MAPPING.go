package msfsrvp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// FSSAGENT_SHARE_MAPPING is the [switch_type(unsigned long)] union of share-mapping
// levels ([MS-FSRVP] 2.2.3.1). Tag carries the discriminant inline, followed by the
// selected arm; the case(1) arm is a [unique] pointer to the level-1 mapping (the IDL
// declares PFSSAGENT_SHARE_MAPPING_1 under pointer_default(unique)). The [default] arm
// is empty.
type FSSAGENT_SHARE_MAPPING struct {
	Tag           ndr.DWORD                 `ndr:"switch"`
	ShareMapping1 *FSSAGENT_SHARE_MAPPING_1 `ndr:"case=1,unique"`
}
