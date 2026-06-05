package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_103 contains information about a server, including capabilities
// ([MS-SRVS] 2.2.4.43). Sv103Name, Sv103Comment and Sv103Userpath are [string]
// wchar_t* fields (pointer_default unique); a nil WSTR is a NULL pointer on the
// wire. Sv103Disc is an IDL LONG and Sv103Hidden an IDL BOOL, both signed
// 32-bit.
type SERVER_INFO_103 struct {
	Sv103PlatformId   ndr.DWORD
	Sv103Name         ndr.WSTR `ndr:"unique"`
	Sv103VersionMajor ndr.DWORD
	Sv103VersionMinor ndr.DWORD
	Sv103Type         ndr.DWORD
	Sv103Comment      ndr.WSTR `ndr:"unique"`
	Sv103Users        ndr.DWORD
	Sv103Disc         int32
	Sv103Hidden       int32
	Sv103Announce     ndr.DWORD
	Sv103Anndelta     ndr.DWORD
	Sv103Licenses     ndr.DWORD
	Sv103Userpath     ndr.WSTR `ndr:"unique"`
	Sv103Capabilities ndr.DWORD
}
