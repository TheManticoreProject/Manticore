package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_PROVIDER_DATA is the non-encapsulated discriminated union ([MS-W32T] 2.2.6)
// selected by W32TIME_PROVIDER_INFO.ulProviderType. The discriminant
// (switch_type(unsigned long) -> 4-byte DWORD) is transmitted inline ahead of the
// selected arm ([C706] 14.3.8). Both arms are [unique] pointers.
type W32TIME_PROVIDER_DATA struct {
	Tag                   ndr.DWORD                       `ndr:"switch"`
	PNtpProviderData      *W32TIME_NTP_PROVIDER_DATA      `ndr:"case=0,unique"`
	PHardwareProviderData *W32TIME_HARDWARE_PROVIDER_DATA `ndr:"case=1,unique"`
}
