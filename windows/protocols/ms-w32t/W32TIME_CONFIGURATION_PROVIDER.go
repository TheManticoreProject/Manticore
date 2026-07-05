package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_CONFIGURATION_PROVIDER ([MS-W32T]).
type W32TIME_CONFIGURATION_PROVIDER struct {
	UlSize              ndr.DWORD
	UlInputProvider     ndr.DWORD
	UlEnabled           ndr.DWORD
	WszDllName          *ndr.WSTR `ndr:"unique"`
	WszProviderName     *ndr.WSTR `ndr:"unique"`
	UlDllNameFlag       ndr.DWORD
	UlProviderNameFlag  ndr.DWORD
	UlInputProviderFlag ndr.DWORD
	UlEnabledFlag       ndr.DWORD
	PProviderConfig     *W32TIME_PROVIDER_CONFIG `ndr:"unique"`
}
