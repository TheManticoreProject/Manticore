package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SESSION_INFO_1 contains the details of a session ([MS-SRVS] 2.2.4.13).
// sesi1_cname and sesi1_username are [string] wchar_t* fields (pointer_default
// unique); a nil WSTR is a NULL pointer on the wire.
type SESSION_INFO_1 struct {
	Sesi1Cname     ndr.WSTR `ndr:"unique"`
	Sesi1Username  ndr.WSTR `ndr:"unique"`
	Sesi1NumOpens  ndr.DWORD
	Sesi1Time      ndr.DWORD
	Sesi1IdleTime  ndr.DWORD
	Sesi1UserFlags ndr.DWORD
}
