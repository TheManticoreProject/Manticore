package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SESSION_ENUM_UNION is the [switch_type(DWORD)] union of session enumeration
// containers ([MS-SRVS] 2.2.4.22). Tag carries the discriminant (the level) inline,
// followed by the selected arm. Each arm is a [unique] pointer to its container.
type SESSION_ENUM_UNION struct {
	Tag      ndr.DWORD                   `ndr:"switch"`
	Level0   *SESSION_INFO_0_CONTAINER   `ndr:"case=0,unique"`
	Level1   *SESSION_INFO_1_CONTAINER   `ndr:"case=1,unique"`
	Level2   *SESSION_INFO_2_CONTAINER   `ndr:"case=2,unique"`
	Level10  *SESSION_INFO_10_CONTAINER  `ndr:"case=10,unique"`
	Level502 *SESSION_INFO_502_CONTAINER `ndr:"case=502,unique"`
}
