package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SESSION_INFO_0_CONTAINER is the level-0 session enumeration container ([MS-SRVS]
// 2.2.4.17). Buffer is a [unique] pointer to a conformant array of SESSION_INFO_0
// sized by EntriesRead.
type SESSION_INFO_0_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []SESSION_INFO_0 `ndr:"unique,size_is=EntriesRead"`
}
