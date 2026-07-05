package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_CONFIGURATION_ADVANCED ([MS-W32T]).
type W32TIME_CONFIGURATION_ADVANCED struct {
	UlSize                     ndr.DWORD
	UlFrequencyCorrectRate     ndr.DWORD
	UlPollAdjustFactor         ndr.DWORD
	UlLargePhaseOffset         ndr.DWORD
	UlSpikeWatchPeriod         ndr.DWORD
	UlLocalClockDispersion     ndr.DWORD
	UlHoldPeriod               ndr.DWORD
	UlPhaseCorrectRate         ndr.DWORD
	UlUpdateInterval           ndr.DWORD
	UlFrequencyCorrectRateFlag ndr.DWORD
	UlPollAdjustFactorFlag     ndr.DWORD
	UlLargePhaseOffsetFlag     ndr.DWORD
	UlSpikeWatchPeriodFlag     ndr.DWORD
	UlLocalClockDispersionFlag ndr.DWORD
	UlHoldPeriodFlag           ndr.DWORD
	UlPhaseCorrectRateFlag     ndr.DWORD
	UlUpdateIntervalFlag       ndr.DWORD
}
