package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SESSION_INFO_0 contains the name of a computer that established a session
// ([MS-SRVS] 2.2.4.12). sesi0_cname is a [string] wchar_t* field (pointer_default
// unique); a nil WSTR is a NULL pointer on the wire.
type SESSION_INFO_0 struct {
	Sesi0Cname ndr.WSTR `ndr:"unique"`
}
