package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_TRANSPORT_INFO_3 contains information about a server transport
// ([MS-SRVS] 2.2.4.97). Svti3Transportaddress is a [unique] pointer to a
// conformant byte array sized by Svti3Transportaddresslength. Svti3Password
// is a fixed 256-byte array. The string fields are [string] wchar_t*
// (pointer_default unique).
type SERVER_TRANSPORT_INFO_3 struct {
	Svti3Numberofvcs            ndr.DWORD
	Svti3Transportname          ndr.WSTR `ndr:"unique"`
	Svti3Transportaddress       []byte   `ndr:"unique,size_is=Svti3Transportaddresslength"`
	Svti3Transportaddresslength ndr.DWORD
	Svti3Networkaddress         ndr.WSTR `ndr:"unique"`
	Svti3Domain                 ndr.WSTR `ndr:"unique"`
	Svti3Flags                  ndr.DWORD
	Svti3Passwordlength         ndr.DWORD
	Svti3Password               [256]byte
}
