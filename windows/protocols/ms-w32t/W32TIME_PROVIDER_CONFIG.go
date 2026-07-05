package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_PROVIDER_CONFIG ([MS-W32T]).
type W32TIME_PROVIDER_CONFIG struct {
	UlSize              ndr.DWORD
	UlProviderType      ndr.DWORD
	PProviderConfigData *W32TIME_PROVIDER_CONFIG_DATA `ndr:"unique"`
}
