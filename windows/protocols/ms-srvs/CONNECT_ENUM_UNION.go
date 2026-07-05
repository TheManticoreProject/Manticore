package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// CONNECT_ENUM_UNION is the [switch_type(DWORD)] union of connection enumeration
// containers ([MS-SRVS] 2.2.4.5). Tag carries the discriminant (the level) inline,
// followed by the selected arm. Each arm is a [unique] pointer to its container.
type CONNECT_ENUM_UNION struct {
	Tag    ndr.DWORD                 `ndr:"switch"`
	Level0 *CONNECT_INFO_0_CONTAINER `ndr:"case=0,unique"`
	Level1 *CONNECT_INFO_1_CONTAINER `ndr:"case=1,unique"`
}
