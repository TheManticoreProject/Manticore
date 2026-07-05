package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_PROVIDER_INFO ([MS-W32T]).
type W32TIME_PROVIDER_INFO struct {
	UlProviderType ndr.DWORD
	ProviderData   W32TIME_PROVIDER_DATA
}
