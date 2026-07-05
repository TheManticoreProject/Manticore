package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_XPORT_INFO_1_CONTAINER is the level-1 server transport enumeration
// container ([MS-SRVS] 2.2.4.99). Buffer is a [unique] pointer to a conformant
// array of SERVER_TRANSPORT_INFO_1 sized by EntriesRead.
type SERVER_XPORT_INFO_1_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []SERVER_TRANSPORT_INFO_1 `ndr:"unique,size_is=EntriesRead"`
}
