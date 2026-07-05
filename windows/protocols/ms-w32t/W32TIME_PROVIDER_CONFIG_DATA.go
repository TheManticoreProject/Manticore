package msw32t

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// W32TIME_PROVIDER_CONFIG_DATA is the non-encapsulated discriminated union
// ([MS-W32T] 2.2.10) selected by W32TIME_PROVIDER_CONFIG.ulProviderType. The
// discriminant (switch_type(unsigned long) -> 4-byte DWORD) is transmitted inline
// ahead of the selected arm ([C706] 14.3.8). Both arms are [unique] pointers in the
// IDL (PW32TIME_NTPCLIENT_PROVIDER_CONFIG_DATA / PW32TIME_NTPSERVER_PROVIDER_CONFIG_DATA).
type W32TIME_PROVIDER_CONFIG_DATA struct {
	Tag                          ndr.DWORD                               `ndr:"switch"`
	PNtpClientProviderConfigData *W32TIME_NTPCLIENT_PROVIDER_CONFIG_DATA `ndr:"case=0,unique"`
	PNtpServerProviderConfigData *W32TIME_NTPSERVER_PROVIDER_CONFIG_DATA `ndr:"case=1,unique"`
}
