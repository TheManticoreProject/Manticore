package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_CONFIGURATION_DEFAULT ([MS-W32T]).
type W32TIME_CONFIGURATION_DEFAULT struct {
	UlSize               ndr.DWORD
	WszFileLogName       *ndr.WSTR `ndr:"unique"`
	WszFileLogEntries    *ndr.WSTR `ndr:"unique"`
	UlFileLogSize        ndr.DWORD
	UlFileLogFlags       ndr.DWORD
	UlFileLogNameFlag    ndr.DWORD
	UlFileLogEntriesFlag ndr.DWORD
	UlFileLogSizeFlag    ndr.DWORD
	UlFileLogFlagsFlag   ndr.DWORD
}
