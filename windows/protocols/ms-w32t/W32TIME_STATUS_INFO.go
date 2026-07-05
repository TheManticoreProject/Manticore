package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_STATUS_INFO ([MS-W32T]).
type W32TIME_STATUS_INFO struct {
	UlSize                ndr.DWORD
	ELeapIndicator        ndr.DWORD
	NStratum              ndr.DWORD
	NPollInterval         int32
	RefidSource           ndr.DWORD
	QwLastSyncTicks       uint64
	ToRootDelay           int64
	TpRootDispersion      uint64
	NClockPrecision       int32
	WszSource             *ndr.WSTR `ndr:"unique"`
	ToSysPhaseOffset      int64
	UlLcState             ndr.DWORD
	UlTSFlags             ndr.DWORD
	UlClockRate           ndr.DWORD
	UlNetlogonServiceBits ndr.DWORD
	ELastSyncResult       ndr.DWORD
	TpTimeLastGoodSync    uint64
	CEntries              ndr.DWORD
	PEntries              []W32TIME_ENTRY `ndr:"unique,size_is=CEntries"`
}
