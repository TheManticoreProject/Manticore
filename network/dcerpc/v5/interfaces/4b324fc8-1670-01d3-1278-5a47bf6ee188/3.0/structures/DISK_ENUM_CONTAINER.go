package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// DISK_ENUM_CONTAINER is the disk enumeration container ([MS-SRVS] 2.2.4.87).
// Buffer is a [unique] pointer to a conformant-varying array of DISK_INFO whose
// maximum and actual counts are both EntriesRead (size_is and length_is).
type DISK_ENUM_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []DISK_INFO `ndr:"unique,varying,size_is=EntriesRead,length_is=EntriesRead"`
}
