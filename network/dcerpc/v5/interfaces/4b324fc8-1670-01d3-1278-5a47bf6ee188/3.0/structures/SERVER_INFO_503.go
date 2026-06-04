package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_503 contains information about a server ([MS-SRVS] 2.2.4.45).
// Sv503Domain is a [string] wchar_t* field (pointer_default unique); a nil WSTR
// is a NULL pointer on the wire. The fields declared as IDL int are signed
// 32-bit; all others are DWORD.
type SERVER_INFO_503 struct {
	Sv503Sessopens               ndr.DWORD
	Sv503Sessvcs                 ndr.DWORD
	Sv503Opensearch              ndr.DWORD
	Sv503Sizreqbuf               ndr.DWORD
	Sv503Initworkitems           ndr.DWORD
	Sv503Maxworkitems            ndr.DWORD
	Sv503Rawworkitems            ndr.DWORD
	Sv503Irpstacksize            ndr.DWORD
	Sv503Maxrawbuflen            ndr.DWORD
	Sv503Sessusers               ndr.DWORD
	Sv503Sessconns               ndr.DWORD
	Sv503Maxpagedmemoryusage     ndr.DWORD
	Sv503Maxnonpagedmemoryusage  ndr.DWORD
	Sv503Enablesoftcompat        int32
	Sv503Enableforcedlogoff      int32
	Sv503Timesource              int32
	Sv503Acceptdownlevelapis     int32
	Sv503Lmannounce              int32
	Sv503Domain                  ndr.WSTR `ndr:"unique"`
	Sv503Maxcopyreadlen          ndr.DWORD
	Sv503Maxcopywritelen         ndr.DWORD
	Sv503Minkeepsearch           ndr.DWORD
	Sv503Maxkeepsearch           ndr.DWORD
	Sv503Minkeepcomplsearch      ndr.DWORD
	Sv503Maxkeepcomplsearch      ndr.DWORD
	Sv503Threadcountadd          ndr.DWORD
	Sv503Numblockthreads         ndr.DWORD
	Sv503Scavtimeout             ndr.DWORD
	Sv503Minrcvqueue             ndr.DWORD
	Sv503Minfreeworkitems        ndr.DWORD
	Sv503Xactmemsize             ndr.DWORD
	Sv503Threadpriority          ndr.DWORD
	Sv503Maxmpxct                ndr.DWORD
	Sv503Oplockbreakwait         ndr.DWORD
	Sv503Oplockbreakresponsewait ndr.DWORD
	Sv503Enableoplocks           int32
	Sv503Enableoplockforceclose  int32
	Sv503Enablefcbopens          int32
	Sv503Enableraw               int32
	Sv503Enablesharednetdrives   int32
	Sv503Minfreeconnections      ndr.DWORD
	Sv503Maxfreeconnections      ndr.DWORD
}
