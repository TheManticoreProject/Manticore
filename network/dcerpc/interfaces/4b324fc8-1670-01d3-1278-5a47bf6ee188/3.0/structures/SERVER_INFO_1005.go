package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1005 contains the server comment ([MS-SRVS] 2.2.4.47).
// Sv1005Comment is a [string] wchar_t* field (pointer_default unique); a nil
// WSTR is a NULL pointer on the wire.
type SERVER_INFO_1005 struct {
	Sv1005Comment ndr.WSTR `ndr:"unique"`
}
