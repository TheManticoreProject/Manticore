package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// FILE_ENUM_UNION is the [switch_type(DWORD)] union of file enumeration containers
// ([MS-SRVS] 2.2.4.10). Tag carries the discriminant (the level) inline, followed
// by the selected arm. Each arm is a [unique] pointer to its container.
type FILE_ENUM_UNION struct {
	Tag    ndr.DWORD              `ndr:"switch"`
	Level2 *FILE_INFO_2_CONTAINER `ndr:"case=2,unique"`
	Level3 *FILE_INFO_3_CONTAINER `ndr:"case=3,unique"`
}
