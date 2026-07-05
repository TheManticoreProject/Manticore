package msw32t

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// wstr returns a pointer to an ndr.WSTR carrying s, for building [unique] string fields.
func wstr(s string) *ndr.WSTR {
	w := ndr.WSTR(s)
	return &w
}

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts deep
// equality — the wire-shape acceptance check for a structure's NDR tags.
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}
	var out T
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("%s: round trip mismatch:\n in:  %+v\n out: %+v", name, in, out)
	}
}

// TestW32TIME_ENTRY_RoundTrip covers a struct of three [unique] wide-string pointers, with
// both the populated and the all-nil cases.
func TestW32TIME_ENTRY_RoundTrip(t *testing.T) {
	roundTrip(t, "W32TIME_ENTRY", W32TIME_ENTRY{
		UlSize:   40,
		WszName:  wstr("NtpServer"),
		WszValue: wstr("time.windows.com,0x9"),
		WszHelp:  wstr("List of peers"),
	})
	roundTrip(t, "W32TIME_ENTRY/nil-strings", W32TIME_ENTRY{UlSize: 8})
}

// TestW32TIME_NTP_PEER_INFO_RoundTrip covers a struct interleaving DWORDs, 64-bit fields, a
// [unique] string, and a run of unsigned char fields.
func TestW32TIME_NTP_PEER_INFO_RoundTrip(t *testing.T) {
	roundTrip(t, "W32TIME_NTP_PEER_INFO", W32TIME_NTP_PEER_INFO{
		UlSize:                72,
		UlResolveAttempts:     3,
		U64TimeRemaining:      0x0000000100000000,
		U64LastSuccessfulSync: 0x01d9a1b2c3d4e5f6,
		UlLastSyncError:       0,
		UlLastSyncErrorMsgId:  0,
		UlValidDataCounter:    12,
		UlAuthTypeMsgId:       1,
		WszUniqueName:         wstr("time.windows.com"),
		UlMode:                3,
		UlStratum:             2,
		UlReachability:        0xFF,
		UlPeerPollInterval:    10,
		UlHostPollInterval:    6,
	})
}

// TestW32TIME_NTP_PROVIDER_DATA_RoundTrip covers a [unique] size_is array of pointer-free
// structs, plus the empty-array case.
func TestW32TIME_NTP_PROVIDER_DATA_RoundTrip(t *testing.T) {
	roundTrip(t, "W32TIME_NTP_PROVIDER_DATA", W32TIME_NTP_PROVIDER_DATA{
		UlSize:       64,
		UlError:      0,
		UlErrorMsgId: 0,
		CPeerInfo:    2,
		PPeerInfo: []W32TIME_NTP_PEER_INFO{
			{UlSize: 72, WszUniqueName: wstr("peer-a"), UlStratum: 2},
			{UlSize: 72, WszUniqueName: wstr("peer-b"), UlStratum: 3},
		},
	})
	roundTrip(t, "W32TIME_NTP_PROVIDER_DATA/empty", W32TIME_NTP_PROVIDER_DATA{UlSize: 16})
}

// TestW32TIME_PROVIDER_INFO_RoundTrip covers the switch_is(ulProviderType) union carried by
// value, with each pointer arm selected in turn.
func TestW32TIME_PROVIDER_INFO_RoundTrip(t *testing.T) {
	roundTrip(t, "W32TIME_PROVIDER_INFO/ntp", W32TIME_PROVIDER_INFO{
		UlProviderType: 0,
		ProviderData: W32TIME_PROVIDER_DATA{
			Tag: 0,
			PNtpProviderData: &W32TIME_NTP_PROVIDER_DATA{
				UlSize:    64,
				CPeerInfo: 1,
				PPeerInfo: []W32TIME_NTP_PEER_INFO{{UlSize: 72, WszUniqueName: wstr("peer")}},
			},
		},
	})
	roundTrip(t, "W32TIME_PROVIDER_INFO/hardware", W32TIME_PROVIDER_INFO{
		UlProviderType: 1,
		ProviderData: W32TIME_PROVIDER_DATA{
			Tag: 1,
			PHardwareProviderData: &W32TIME_HARDWARE_PROVIDER_DATA{
				UlSize:                 32,
				WszReferenceIdentifier: wstr("GPS"),
			},
		},
	})
}

// TestW32TIME_PROVIDER_CONFIG_RoundTrip covers a [unique] pointer to the switch_is union
// whose arms are themselves [unique] pointers, for both cases.
func TestW32TIME_PROVIDER_CONFIG_RoundTrip(t *testing.T) {
	roundTrip(t, "W32TIME_PROVIDER_CONFIG/client", W32TIME_PROVIDER_CONFIG{
		UlSize:         48,
		UlProviderType: 0,
		PProviderConfigData: &W32TIME_PROVIDER_CONFIG_DATA{
			Tag: 0,
			PNtpClientProviderConfigData: &W32TIME_NTPCLIENT_PROVIDER_CONFIG_DATA{
				UlSize:                48,
				UlSpecialPollInterval: 3600,
				WszType:               wstr("NTP"),
				WszNtpServer:          wstr("time.windows.com,0x9"),
				CEntries:              1,
				PEntries:              []W32TIME_ENTRY{{UlSize: 16, WszName: wstr("k")}},
			},
		},
	})
	roundTrip(t, "W32TIME_PROVIDER_CONFIG/server", W32TIME_PROVIDER_CONFIG{
		UlSize:         48,
		UlProviderType: 1,
		PProviderConfigData: &W32TIME_PROVIDER_CONFIG_DATA{
			Tag: 1,
			PNtpServerProviderConfigData: &W32TIME_NTPSERVER_PROVIDER_CONFIG_DATA{
				UlSize:          24,
				UlEventLogFlags: 2,
				CEntries:        0,
			},
		},
	})
	roundTrip(t, "W32TIME_PROVIDER_CONFIG/nil", W32TIME_PROVIDER_CONFIG{UlSize: 12})
}

// TestW32TIME_CONFIGURATION_INFO_RoundTrip covers the top-level configuration blob: three
// nested value structs, a size_is array of [unique] pointers, and a size_is array of structs.
func TestW32TIME_CONFIGURATION_INFO_RoundTrip(t *testing.T) {
	roundTrip(t, "W32TIME_CONFIGURATION_INFO", W32TIME_CONFIGURATION_INFO{
		UlSize:          200,
		BasicConfig:     W32TIME_CONFIGURATION_BASIC{UlSize: 68, UlAnnounceFlags: 0x0A, UlMinPollInterval: 6},
		AdvancedConfig:  W32TIME_CONFIGURATION_ADVANCED{UlSize: 68, UlPollAdjustFactor: 5},
		DefaultConfig:   W32TIME_CONFIGURATION_DEFAULT{UlSize: 36, WszFileLogName: wstr("C:\\w32tm.log")},
		CProviderConfig: 2,
		PProviderConfig: []*W32TIME_CONFIGURATION_PROVIDER{
			{UlSize: 40, UlEnabled: 1, WszProviderName: wstr("NtpClient")},
			{UlSize: 40, UlEnabled: 1, WszProviderName: wstr("NtpServer")},
		},
		CEntries: 1,
		PEntries: []W32TIME_ENTRY{{UlSize: 16, WszName: wstr("Type"), WszValue: wstr("NTP")}},
	})
	roundTrip(t, "W32TIME_CONFIGURATION_INFO/empty", W32TIME_CONFIGURATION_INFO{UlSize: 40})
}

// TestW32TIME_STATUS_INFO_RoundTrip covers the status blob mixing signed/unsigned 64-bit
// fields, a [unique] string, and a size_is array of entries.
func TestW32TIME_STATUS_INFO_RoundTrip(t *testing.T) {
	roundTrip(t, "W32TIME_STATUS_INFO", W32TIME_STATUS_INFO{
		UlSize:             128,
		ELeapIndicator:     0,
		NStratum:           3,
		NPollInterval:      -6,
		RefidSource:        0x4C4F434C,
		QwLastSyncTicks:    0x01d9a1b2c3d4e5f6,
		ToRootDelay:        -12345,
		TpRootDispersion:   6789,
		NClockPrecision:    -20,
		WszSource:          wstr("time.windows.com"),
		ToSysPhaseOffset:   -42,
		UlLcState:          2,
		ELastSyncResult:    0,
		TpTimeLastGoodSync: 0x01d9a1b200000000,
		CEntries:           1,
		PEntries:           []W32TIME_ENTRY{{UlSize: 16, WszName: wstr("LastSync")}},
	})
	roundTrip(t, "W32TIME_STATUS_INFO/nil-source", W32TIME_STATUS_INFO{UlSize: 64, NStratum: 1})
}
