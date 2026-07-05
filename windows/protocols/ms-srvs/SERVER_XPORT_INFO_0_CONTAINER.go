package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_XPORT_INFO_0_CONTAINER is the level-0 server transport enumeration
// container ([MS-SRVS] 2.2.4.98). Buffer is a [unique] pointer to a conformant
// array of SERVER_TRANSPORT_INFO_0 sized by EntriesRead.
type SERVER_XPORT_INFO_0_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []SERVER_TRANSPORT_INFO_0 `ndr:"unique,size_is=EntriesRead"`
}
