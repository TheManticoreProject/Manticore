package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO_1005 contains the flags of a shared resource ([MS-SRVS] 2.2.4.29).
type SHARE_INFO_1005 struct {
	Shi1005Flags ndr.DWORD
}
