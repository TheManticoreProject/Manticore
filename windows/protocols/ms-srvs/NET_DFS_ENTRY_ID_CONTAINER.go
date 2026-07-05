package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// NET_DFS_ENTRY_ID_CONTAINER is the DFS entry-id container ([MS-SRVS]
// 2.2.4.89). Buffer is a [unique] pointer to a conformant array of
// NET_DFS_ENTRY_ID sized by Count.
type NET_DFS_ENTRY_ID_CONTAINER struct {
	Count  ndr.DWORD
	Buffer []NET_DFS_ENTRY_ID `ndr:"unique,size_is=Count"`
}
