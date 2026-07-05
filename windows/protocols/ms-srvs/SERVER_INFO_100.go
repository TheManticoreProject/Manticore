package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_100 contains information about a server ([MS-SRVS] 2.2.4.40).
// It is referenced by the SERVER_INFO union (case 100) but is defined in the
// specification rather than the interface IDL. Sv100Name is a [string] wchar_t*
// field (pointer_default unique); a nil WSTR is a NULL pointer on the wire.
type SERVER_INFO_100 struct {
	Sv100PlatformId ndr.DWORD
	Sv100Name       ndr.WSTR `ndr:"unique"`
}
