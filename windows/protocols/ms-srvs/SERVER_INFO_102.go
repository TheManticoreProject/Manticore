package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_102 contains information about a server ([MS-SRVS] 2.2.4.42).
// Sv102Name, Sv102Comment and Sv102Userpath are [string] wchar_t* fields
// (pointer_default unique); a nil WSTR is a NULL pointer on the wire. Sv102Disc
// is an IDL long and Sv102Hidden an IDL int, both signed 32-bit.
type SERVER_INFO_102 struct {
	Sv102PlatformId   ndr.DWORD
	Sv102Name         ndr.WSTR `ndr:"unique"`
	Sv102VersionMajor ndr.DWORD
	Sv102VersionMinor ndr.DWORD
	Sv102Type         ndr.DWORD
	Sv102Comment      ndr.WSTR `ndr:"unique"`
	Sv102Users        ndr.DWORD
	Sv102Disc         int32
	Sv102Hidden       int32
	Sv102Announce     ndr.DWORD
	Sv102Anndelta     ndr.DWORD
	Sv102Licenses     ndr.DWORD
	Sv102Userpath     ndr.WSTR `ndr:"unique"`
}
