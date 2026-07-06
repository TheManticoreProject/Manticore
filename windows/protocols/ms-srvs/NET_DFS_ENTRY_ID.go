package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// NET_DFS_ENTRY_ID identifies a DFS root or link by GUID and prefix ([MS-SRVS]
// 2.2.4.88). Uid is an inline GUID value; Prefix is a [string] WCHAR*
// (pointer_default unique).
type NET_DFS_ENTRY_ID struct {
	Uid    msdtyp.GUID
	Prefix ndr.WSTR `ndr:"unique"`
}
