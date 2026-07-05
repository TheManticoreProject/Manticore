package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO_2 contains information about a shared resource ([MS-SRVS] 2.2.4.24).
// shi2_netname, shi2_remark, shi2_path and shi2_passwd are embedded [unique] [string]
// wchar_t*.
type SHARE_INFO_2 struct {
	Shi2Netname     ndr.WSTR `ndr:"unique"`
	Shi2Type        ndr.DWORD
	Shi2Remark      ndr.WSTR `ndr:"unique"`
	Shi2Permissions ndr.DWORD
	Shi2MaxUses     ndr.DWORD
	Shi2CurrentUses ndr.DWORD
	Shi2Path        ndr.WSTR `ndr:"unique"`
	Shi2Passwd      ndr.WSTR `ndr:"unique"`
}
