package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_TRANSPORT_INFO_0 contains information about a server transport
// ([MS-SRVS] 2.2.4.94). Svti0Transportaddress is a [unique] pointer to a
// conformant byte array sized by Svti0Transportaddresslength. The string
// fields are [string] wchar_t* (pointer_default unique).
type SERVER_TRANSPORT_INFO_0 struct {
	Svti0Numberofvcs            ndr.DWORD
	Svti0Transportname          ndr.WSTR `ndr:"unique"`
	Svti0Transportaddress       []byte   `ndr:"unique,size_is=Svti0Transportaddresslength"`
	Svti0Transportaddresslength ndr.DWORD
	Svti0Networkaddress         ndr.WSTR `ndr:"unique"`
}
