package mscmpo

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// BIND_INFO_BLOB ([MS-CMPO] 2.2.3.1): the fixed 8-byte payload carried in the rguchBlob
// byte arrays of Poke/BuildContext. grbitComProtocols is a COM_PROTOCOL bitmask of the
// RPC protocol sequences the caller supports.
type BIND_INFO_BLOB struct {
	DwcbThisStruct    ndr.DWORD
	GrbitComProtocols COM_PROTOCOL
}
