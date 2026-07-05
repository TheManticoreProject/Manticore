package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SESSION_INFO_502 contains the full details of a session including the transport
// ([MS-SRVS] 2.2.4.16). sesi502_cname, sesi502_username, sesi502_cltype_name, and
// sesi502_transport are [string] wchar_t* fields (pointer_default unique); a nil
// WSTR is a NULL pointer on the wire.
type SESSION_INFO_502 struct {
	Sesi502Cname      ndr.WSTR `ndr:"unique"`
	Sesi502Username   ndr.WSTR `ndr:"unique"`
	Sesi502NumOpens   ndr.DWORD
	Sesi502Time       ndr.DWORD
	Sesi502IdleTime   ndr.DWORD
	Sesi502UserFlags  ndr.DWORD
	Sesi502CltypeName ndr.WSTR `ndr:"unique"`
	Sesi502Transport  ndr.WSTR `ndr:"unique"`
}
