package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_NTPSERVER_PROVIDER_CONFIG_DATA ([MS-W32T]).
type W32TIME_NTPSERVER_PROVIDER_CONFIG_DATA struct {
	UlSize                                 ndr.DWORD
	UlAllowNonstandardModeCombinations     ndr.DWORD
	UlAllowNonstandardModeCombinationsFlag ndr.DWORD
	UlEventLogFlags                        ndr.DWORD
	UlEventLogFlagsFlag                    ndr.DWORD
	CEntries                               ndr.DWORD
	PEntries                               []W32TIME_ENTRY `ndr:"unique,size_is=CEntries"`
}
