package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_XPORT_ENUM_UNION is the [switch_type(DWORD)] union of server transport
// enumeration containers ([MS-SRVS] 2.2.4.102). Tag carries the discriminant
// (the level) inline, followed by the selected arm. Each arm is a [unique]
// pointer to its container.
type SERVER_XPORT_ENUM_UNION struct {
	Tag    ndr.DWORD                      `ndr:"switch"`
	Level0 *SERVER_XPORT_INFO_0_CONTAINER `ndr:"case=0,unique"`
	Level1 *SERVER_XPORT_INFO_1_CONTAINER `ndr:"case=1,unique"`
	Level2 *SERVER_XPORT_INFO_2_CONTAINER `ndr:"case=2,unique"`
	Level3 *SERVER_XPORT_INFO_3_CONTAINER `ndr:"case=3,unique"`
}
