package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// NET_DFS_ENTRY_ID identifies a DFS root or link by GUID and prefix ([MS-SRVS]
// 2.2.4.88). Uid is an inline GUID value; Prefix is a [string] WCHAR*
// (pointer_default unique).
type NET_DFS_ENTRY_ID struct {
	Uid    dtyp.GUID
	Prefix ndr.WSTR `ndr:"unique"`
}
