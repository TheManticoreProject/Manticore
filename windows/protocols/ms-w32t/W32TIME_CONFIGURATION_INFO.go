package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_CONFIGURATION_INFO ([MS-W32T]).
type W32TIME_CONFIGURATION_INFO struct {
	UlSize          ndr.DWORD
	BasicConfig     W32TIME_CONFIGURATION_BASIC
	AdvancedConfig  W32TIME_CONFIGURATION_ADVANCED
	DefaultConfig   W32TIME_CONFIGURATION_DEFAULT
	CProviderConfig ndr.DWORD
	PProviderConfig []*W32TIME_CONFIGURATION_PROVIDER `ndr:"elem=unique,unique,size_is=CProviderConfig"`
	CEntries        ndr.DWORD
	PEntries        []W32TIME_ENTRY `ndr:"unique,size_is=CEntries"`
}
