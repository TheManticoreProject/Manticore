package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// CONNECTION_INFO_1 contains the details of a connection ([MS-SRVS] 2.2.4.2).
// coni1_username and coni1_netname are [string] wchar_t* fields (pointer_default
// unique); a nil WSTR is a NULL pointer on the wire.
type CONNECTION_INFO_1 struct {
	Coni1Id       ndr.DWORD
	Coni1Type     ndr.DWORD
	Coni1NumOpens ndr.DWORD
	Coni1NumUsers ndr.DWORD
	Coni1Time     ndr.DWORD
	Coni1Username ndr.WSTR `ndr:"unique"`
	Coni1Netname  ndr.WSTR `ndr:"unique"`
}
