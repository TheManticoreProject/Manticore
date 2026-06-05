package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_101 contains information about a server ([MS-SRVS] 2.2.4.41).
// It is referenced by the SERVER_INFO union (case 101) but is defined in the
// specification rather than the interface IDL. Sv101Name and Sv101Comment are
// [string] wchar_t* fields (pointer_default unique); a nil WSTR is a NULL
// pointer on the wire.
type SERVER_INFO_101 struct {
	Sv101PlatformId   ndr.DWORD
	Sv101Name         ndr.WSTR `ndr:"unique"`
	Sv101VersionMajor ndr.DWORD
	Sv101VersionMinor ndr.DWORD
	Sv101Type         ndr.DWORD
	Sv101Comment      ndr.WSTR `ndr:"unique"`
}
