package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SESSION_INFO_2 contains the details of a session including the client type
// ([MS-SRVS] 2.2.4.14). sesi2_cname, sesi2_username, and sesi2_cltype_name are
// [string] wchar_t* fields (pointer_default unique); a nil WSTR is a NULL pointer.
type SESSION_INFO_2 struct {
	Sesi2Cname      ndr.WSTR `ndr:"unique"`
	Sesi2Username   ndr.WSTR `ndr:"unique"`
	Sesi2NumOpens   ndr.DWORD
	Sesi2Time       ndr.DWORD
	Sesi2IdleTime   ndr.DWORD
	Sesi2UserFlags  ndr.DWORD
	Sesi2CltypeName ndr.WSTR `ndr:"unique"`
}
