package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_502 contains information about a server ([MS-SRVS] 2.2.4.44).
// The five trailing fields are IDL int (signed 32-bit); all others are DWORD.
type SERVER_INFO_502 struct {
	Sv502Sessopens              ndr.DWORD
	Sv502Sessvcs                ndr.DWORD
	Sv502Opensearch             ndr.DWORD
	Sv502Sizreqbuf              ndr.DWORD
	Sv502Initworkitems          ndr.DWORD
	Sv502Maxworkitems           ndr.DWORD
	Sv502Rawworkitems           ndr.DWORD
	Sv502Irpstacksize           ndr.DWORD
	Sv502Maxrawbuflen           ndr.DWORD
	Sv502Sessusers              ndr.DWORD
	Sv502Sessconns              ndr.DWORD
	Sv502Maxpagedmemoryusage    ndr.DWORD
	Sv502Maxnonpagedmemoryusage ndr.DWORD
	Sv502Enablesoftcompat       int32
	Sv502Enableforcedlogoff     int32
	Sv502Timesource             int32
	Sv502Acceptdownlevelapis    int32
	Sv502Lmannounce             int32
}
