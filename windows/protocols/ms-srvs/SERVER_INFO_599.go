package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_599 contains information about a server ([MS-SRVS] 2.2.4.46).
// Sv599Domain is a [string] wchar_t* field (pointer_default unique); a nil WSTR
// is a NULL pointer on the wire. The fields declared as IDL int are signed
// 32-bit; all others are DWORD.
type SERVER_INFO_599 struct {
	Sv599Sessopens               ndr.DWORD
	Sv599Sessvcs                 ndr.DWORD
	Sv599Opensearch              ndr.DWORD
	Sv599Sizreqbuf               ndr.DWORD
	Sv599Initworkitems           ndr.DWORD
	Sv599Maxworkitems            ndr.DWORD
	Sv599Rawworkitems            ndr.DWORD
	Sv599Irpstacksize            ndr.DWORD
	Sv599Maxrawbuflen            ndr.DWORD
	Sv599Sessusers               ndr.DWORD
	Sv599Sessconns               ndr.DWORD
	Sv599Maxpagedmemoryusage     ndr.DWORD
	Sv599Maxnonpagedmemoryusage  ndr.DWORD
	Sv599Enablesoftcompat        int32
	Sv599Enableforcedlogoff      int32
	Sv599Timesource              int32
	Sv599Acceptdownlevelapis     int32
	Sv599Lmannounce              int32
	Sv599Domain                  ndr.WSTR `ndr:"unique"`
	Sv599Maxcopyreadlen          ndr.DWORD
	Sv599Maxcopywritelen         ndr.DWORD
	Sv599Minkeepsearch           ndr.DWORD
	Sv599Maxkeepsearch           ndr.DWORD
	Sv599Minkeepcomplsearch      ndr.DWORD
	Sv599Maxkeepcomplsearch      ndr.DWORD
	Sv599Threadcountadd          ndr.DWORD
	Sv599Numblockthreads         ndr.DWORD
	Sv599Scavtimeout             ndr.DWORD
	Sv599Minrcvqueue             ndr.DWORD
	Sv599Minfreeworkitems        ndr.DWORD
	Sv599Xactmemsize             ndr.DWORD
	Sv599Threadpriority          ndr.DWORD
	Sv599Maxmpxct                ndr.DWORD
	Sv599Oplockbreakwait         ndr.DWORD
	Sv599Oplockbreakresponsewait ndr.DWORD
	Sv599Enableoplocks           int32
	Sv599Enableoplockforceclose  int32
	Sv599Enablefcbopens          int32
	Sv599Enableraw               int32
	Sv599Enablesharednetdrives   int32
	Sv599Minfreeconnections      ndr.DWORD
	Sv599Maxfreeconnections      ndr.DWORD
	Sv599Initsesstable           ndr.DWORD
	Sv599Initconntable           ndr.DWORD
	Sv599Initfiletable           ndr.DWORD
	Sv599Initsearchtable         ndr.DWORD
	Sv599Alertschedule           ndr.DWORD
	Sv599Errorthreshold          ndr.DWORD
	Sv599Networkerrorthreshold   ndr.DWORD
	Sv599Diskspacethreshold      ndr.DWORD
	Sv599Reserved                ndr.DWORD
	Sv599Maxlinkdelay            ndr.DWORD
	Sv599Minlinkthroughput       ndr.DWORD
	Sv599Linkinfovalidtime       ndr.DWORD
	Sv599Scavqosinfoupdatetime   ndr.DWORD
	Sv599Maxworkitemidletime     ndr.DWORD
}
