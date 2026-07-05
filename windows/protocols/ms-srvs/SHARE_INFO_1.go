package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO_1 contains information about a shared resource: name, type and remark
// ([MS-SRVS] 2.2.4.23). shi1_netname and shi1_remark are embedded [unique] [string]
// wchar_t*.
type SHARE_INFO_1 struct {
	Shi1Netname ndr.WSTR `ndr:"unique"`
	Shi1Type    ndr.DWORD
	Shi1Remark  ndr.WSTR `ndr:"unique"`
}
