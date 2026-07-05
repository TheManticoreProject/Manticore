package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_HARDWARE_PROVIDER_DATA ([MS-W32T]).
type W32TIME_HARDWARE_PROVIDER_DATA struct {
	UlSize                 ndr.DWORD
	UlError                ndr.DWORD
	UlErrorMsgId           ndr.DWORD
	WszReferenceIdentifier *ndr.WSTR `ndr:"unique"`
}
