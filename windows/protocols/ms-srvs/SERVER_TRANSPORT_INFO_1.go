package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_TRANSPORT_INFO_1 contains information about a server transport
// ([MS-SRVS] 2.2.4.95). Svti1Transportaddress is a [unique] pointer to a
// conformant byte array sized by Svti1Transportaddresslength. The string
// fields are [string] wchar_t* (pointer_default unique).
type SERVER_TRANSPORT_INFO_1 struct {
	Svti1Numberofvcs            ndr.DWORD
	Svti1Transportname          ndr.WSTR `ndr:"unique"`
	Svti1Transportaddress       []byte   `ndr:"unique,size_is=Svti1Transportaddresslength"`
	Svti1Transportaddresslength ndr.DWORD
	Svti1Networkaddress         ndr.WSTR `ndr:"unique"`
	Svti1Domain                 ndr.WSTR `ndr:"unique"`
}
