package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_NTP_PEER_INFO ([MS-W32T]).
type W32TIME_NTP_PEER_INFO struct {
	UlSize                ndr.DWORD
	UlResolveAttempts     ndr.DWORD
	U64TimeRemaining      uint64
	U64LastSuccessfulSync uint64
	UlLastSyncError       ndr.DWORD
	UlLastSyncErrorMsgId  ndr.DWORD
	UlValidDataCounter    ndr.DWORD
	UlAuthTypeMsgId       ndr.DWORD
	WszUniqueName         *ndr.WSTR `ndr:"unique"`
	UlMode                uint8
	UlStratum             uint8
	UlReachability        uint8
	UlPeerPollInterval    uint8
	UlHostPollInterval    uint8
}
