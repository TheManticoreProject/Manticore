package msbrwsa

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_100 contains basic information about a server ([MS-DTYP] 2.3.11, referenced
// by [MS-BRWSA] 3.1.4.1). It is the element type of SERVER_INFO_100_CONTAINER.Buffer.
// Sv100Name is a [string] wchar_t* field (pointer_default unique); a nil WSTR is a NULL
// pointer on the wire.
type SERVER_INFO_100 struct {
	Sv100PlatformId ndr.DWORD
	Sv100Name       ndr.WSTR `ndr:"unique"`
}
