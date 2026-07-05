package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO_0 contains the name of a shared resource ([MS-SRVS] 2.2.4.22).
// shi0_netname is an embedded [unique] [string] wchar_t*.
type SHARE_INFO_0 struct {
	Shi0Netname ndr.WSTR `ndr:"unique"`
}
