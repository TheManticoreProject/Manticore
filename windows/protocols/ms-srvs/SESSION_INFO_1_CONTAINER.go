package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SESSION_INFO_1_CONTAINER is the level-1 session enumeration container ([MS-SRVS]
// 2.2.4.18). Buffer is a [unique] pointer to a conformant array of SESSION_INFO_1
// sized by EntriesRead.
type SESSION_INFO_1_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []SESSION_INFO_1 `ndr:"unique,size_is=EntriesRead"`
}
