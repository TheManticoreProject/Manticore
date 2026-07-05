package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// GROUP_ATTRIBUTE_INFORMATION contains group fields ([MS-SAMR] 2.2.5.4). It carries the
// attributes (a set of SE_GROUP flags) of a group.
type GROUP_ATTRIBUTE_INFORMATION struct {
	Attributes ndr.DWORD
}
