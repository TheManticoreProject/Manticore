package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO_1006 contains the maximum number of concurrent connections allowed to a
// shared resource ([MS-SRVS] 2.2.4.30).
type SHARE_INFO_1006 struct {
	Shi1006MaxUses ndr.DWORD
}
