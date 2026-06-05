package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SESSION_INFO_10 contains the details of a session ([MS-SRVS] 2.2.4.15).
// sesi10_cname and sesi10_username are [string] wchar_t* fields (pointer_default
// unique); a nil WSTR is a NULL pointer on the wire.
type SESSION_INFO_10 struct {
	Sesi10Cname    ndr.WSTR `ndr:"unique"`
	Sesi10Username ndr.WSTR `ndr:"unique"`
	Sesi10Time     ndr.DWORD
	Sesi10IdleTime ndr.DWORD
}
