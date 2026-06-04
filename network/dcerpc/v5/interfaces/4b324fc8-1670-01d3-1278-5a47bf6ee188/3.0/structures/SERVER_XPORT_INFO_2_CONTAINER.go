package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_XPORT_INFO_2_CONTAINER is the level-2 server transport enumeration
// container ([MS-SRVS] 2.2.4.100). Buffer is a [unique] pointer to a conformant
// array of SERVER_TRANSPORT_INFO_2 sized by EntriesRead.
type SERVER_XPORT_INFO_2_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []SERVER_TRANSPORT_INFO_2 `ndr:"unique,size_is=EntriesRead"`
}
