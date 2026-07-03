package msmsrp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// MSG_INFO_0 holds a single message alias name ([MS-MSRP] 2.2.2.1). msgi0_name is a
// [string] wchar_t* (unique under pointer_default(unique)).
type MSG_INFO_0 struct {
	Msgi0_name *ndr.WSTR `ndr:"unique"`
}
