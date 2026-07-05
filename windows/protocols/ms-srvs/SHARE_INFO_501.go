package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO_501 contains information about a shared resource, including flags
// ([MS-SRVS] 2.2.4.25). shi501_netname and shi501_remark are embedded [unique]
// [string] wchar_t*.
type SHARE_INFO_501 struct {
	Shi501Netname ndr.WSTR `ndr:"unique"`
	Shi501Type    ndr.DWORD
	Shi501Remark  ndr.WSTR `ndr:"unique"`
	Shi501Flags   ndr.DWORD
}
