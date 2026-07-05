package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_NTPCLIENT_PROVIDER_CONFIG_DATA ([MS-W32T]).
type W32TIME_NTPCLIENT_PROVIDER_CONFIG_DATA struct {
	UlSize                                 ndr.DWORD
	UlAllowNonstandardModeCombinations     ndr.DWORD
	UlCrossSiteSyncFlags                   ndr.DWORD
	UlResolvePeerBackoffMinutes            ndr.DWORD
	UlResolvePeerBackoffMaxTimes           ndr.DWORD
	UlCompatibilityFlags                   ndr.DWORD
	UlEventLogFlags                        ndr.DWORD
	UlLargeSampleSkew                      ndr.DWORD
	UlSpecialPollInterval                  ndr.DWORD
	WszType                                *ndr.WSTR `ndr:"unique"`
	WszNtpServer                           *ndr.WSTR `ndr:"unique"`
	UlAllowNonstandardModeCombinationsFlag ndr.DWORD
	UlCrossSiteSyncFlagsFlag               ndr.DWORD
	UlResolvePeerBackoffMinutesFlag        ndr.DWORD
	UlResolvePeerBackoffMaxTimesFlag       ndr.DWORD
	UlCompatibilityFlagsFlag               ndr.DWORD
	UlEventLogFlagsFlag                    ndr.DWORD
	UlLargeSampleSkewFlag                  ndr.DWORD
	UlSpecialPollIntervalFlag              ndr.DWORD
	UlTypeFlag                             ndr.DWORD
	UlNtpServerFlag                        ndr.DWORD
	CEntries                               ndr.DWORD
	PEntries                               []W32TIME_ENTRY `ndr:"unique,size_is=CEntries"`
}
