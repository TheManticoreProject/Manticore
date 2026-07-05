package mstsgu

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TSG_CAPABILITIES_UNION is the [switch_type(unsigned long)] union of capability
// structures ([MS-TSGU] 2.2.9.2.1.1). Tag carries the discriminant (capabilityType)
// inline, followed by the selected arm ([C706] 14.3.8). The single defined arm is a
// value (not a pointer), as declared in the IDL.
type TSG_CAPABILITIES_UNION struct {
	Tag ndr.DWORD `ndr:"switch"`
	// case TSG_CAPABILITY_TYPE_NAP (0x00000001)
	TSGCapNap TSG_CAPABILITY_NAP `ndr:"case=0x00000001"`
}
