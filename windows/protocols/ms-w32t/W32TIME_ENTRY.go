package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_ENTRY ([MS-W32T]).
type W32TIME_ENTRY struct {
	UlSize   ndr.DWORD
	WszName  *ndr.WSTR `ndr:"unique"`
	WszValue *ndr.WSTR `ndr:"unique"`
	WszHelp  *ndr.WSTR `ndr:"unique"`
}
