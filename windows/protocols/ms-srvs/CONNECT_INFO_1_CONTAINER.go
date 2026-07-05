package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// CONNECT_INFO_1_CONTAINER is the level-1 connection enumeration container
// ([MS-SRVS] 2.2.4.4). Buffer is a [unique] pointer to a conformant array of
// CONNECTION_INFO_1 sized by EntriesRead.
type CONNECT_INFO_1_CONTAINER struct {
	EntriesRead ndr.DWORD
	Buffer      []CONNECTION_INFO_1 `ndr:"unique,size_is=EntriesRead"`
}
