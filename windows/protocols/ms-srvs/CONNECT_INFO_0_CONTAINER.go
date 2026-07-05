package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// CONNECT_INFO_0_CONTAINER is the level-0 connection enumeration container
// ([MS-SRVS] 2.2.4.3). Buffer is a [unique] pointer to a conformant array of
// CONNECTION_INFO_0 sized by EntriesRead.
type CONNECT_INFO_0_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []CONNECTION_INFO_0 `ndr:"unique,size_is=EntriesRead"`
}
