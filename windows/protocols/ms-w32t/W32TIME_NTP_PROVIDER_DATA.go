package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_NTP_PROVIDER_DATA ([MS-W32T]).
type W32TIME_NTP_PROVIDER_DATA struct {
	UlSize       ndr.DWORD
	UlError      ndr.DWORD
	UlErrorMsgId ndr.DWORD
	CPeerInfo    ndr.DWORD
	PPeerInfo    []W32TIME_NTP_PEER_INFO `ndr:"unique,size_is=CPeerInfo"`
}
