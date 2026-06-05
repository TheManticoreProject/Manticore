package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO_1004 contains a comment about a shared resource ([MS-SRVS] 2.2.4.28).
// shi1004_remark is an embedded [unique] [string] wchar_t*.
type SHARE_INFO_1004 struct {
	Shi1004Remark ndr.WSTR `ndr:"unique"`
}
