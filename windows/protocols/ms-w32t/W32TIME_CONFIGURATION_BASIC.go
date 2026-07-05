package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_CONFIGURATION_BASIC ([MS-W32T]).
type W32TIME_CONFIGURATION_BASIC struct {
	UlSize                      ndr.DWORD
	UlEventLogFlags             ndr.DWORD
	UlAnnounceFlags             ndr.DWORD
	UlTimeJumpAuditOffset       ndr.DWORD
	UlMinPollInterval           ndr.DWORD
	UlMaxPollInterval           ndr.DWORD
	UlMaxNegPhaseCorrection     ndr.DWORD
	UlMaxPosPhaseCorrection     ndr.DWORD
	UlMaxAllowedPhaseOffset     ndr.DWORD
	UlEventLogFlagsFlag         ndr.DWORD
	UlAnnounceFlagsFlag         ndr.DWORD
	UlTimeJumpAuditOffsetFlag   ndr.DWORD
	UlMinPollIntervalFlag       ndr.DWORD
	UlMaxPollIntervalFlag       ndr.DWORD
	UlMaxNegPhaseCorrectionFlag ndr.DWORD
	UlMaxPosPhaseCorrectionFlag ndr.DWORD
	UlMaxAllowedPhaseOffsetFlag ndr.DWORD
}
