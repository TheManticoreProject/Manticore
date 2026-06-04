package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// FILE_INFO_3_CONTAINER is the level-3 file enumeration container ([MS-SRVS]
// 2.2.4.9). Buffer is a [unique] pointer to a conformant array of FILE_INFO_3
// sized by EntriesRead.
type FILE_INFO_3_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []FILE_INFO_3 `ndr:"unique,size_is=EntriesRead"`
}
