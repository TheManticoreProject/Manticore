package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// FILE_INFO_2_CONTAINER is the level-2 file enumeration container ([MS-SRVS]
// 2.2.4.8). Buffer is a [unique] pointer to a conformant array of FILE_INFO_2
// sized by EntriesRead.
type FILE_INFO_2_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []FILE_INFO_2 `ndr:"unique,size_is=EntriesRead"`
}
