package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_TRANSPORT_INFO_2 contains information about a server transport
// ([MS-SRVS] 2.2.4.96). Svti2Transportaddress is a [unique] pointer to a
// conformant byte array sized by Svti2Transportaddresslength. The string
// fields are [string] wchar_t* (pointer_default unique).
type SERVER_TRANSPORT_INFO_2 struct {
	Svti2Numberofvcs            ndr.DWORD
	Svti2Transportname          ndr.WSTR `ndr:"unique"`
	Svti2Transportaddress       []byte   `ndr:"unique,size_is=Svti2Transportaddresslength"`
	Svti2Transportaddresslength ndr.DWORD
	Svti2Networkaddress         ndr.WSTR `ndr:"unique"`
	Svti2Domain                 ndr.WSTR `ndr:"unique"`
	Svti2Flags                  ndr.DWORD
}
