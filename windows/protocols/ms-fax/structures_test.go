package msfax

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts
// deep equality. This is the wire-shape acceptance gate for the MS-FAX NDR structures
// in the absence of a live Windows Fax service (the interface is reached over the
// SHAREDFAX named pipe, which cannot be driven from unit tests). Note that most FAX
// Get*/Enum* methods return their structures custom-marshaled into opaque byte buffers
// ([MS-FAX] section 2.2); these round-trip tests cover the standalone NDR layout of the
// types themselves.
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

func wstr(s string) *ndr.WSTR { w := ndr.WSTR(s); return &w }

// TestEnumWidths verifies the FAX enums are transmitted as 16-bit NDR enums
// ([C706] 14.3.6): a uint32-backed enum would silently emit 4 bytes and corrupt the wire.
func TestEnumWidths(t *testing.T) {
	cases := []struct {
		name string
		v    any
	}{
		{"FAX_ENUM_MESSAGE_FOLDER", struct{ V FAX_ENUM_MESSAGE_FOLDER }{FAX_MESSAGE_FOLDER_QUEUE}},
		{"FAX_ENUM_GROUP_STATUS", struct{ V FAX_ENUM_GROUP_STATUS }{FAX_GROUP_STATUS_EMPTY}},
		{"FAX_ENUM_PRIORITY_TYPE", struct{ V FAX_ENUM_PRIORITY_TYPE }{FAX_PRIORITY_TYPE_HIGH}},
		{"PRODUCT_SKU_TYPE", struct{ V PRODUCT_SKU_TYPE }{PRODUCT_SKU_SERVER}},
		{"FAX_ENUM_CONFIG_OPTION", struct{ V FAX_ENUM_CONFIG_OPTION }{FAX_CONFIG_OPTION_QUEUE_STATE}},
	}
	for _, c := range cases {
		raw, err := ndr.Marshal(&c.v)
		if err != nil {
			t.Fatalf("%s: Marshal: %v", c.name, err)
		}
		if len(raw) != 2 {
			t.Errorf("%s marshalled to %d bytes, want 2 (16-bit NDR enum)", c.name, len(raw))
		}
	}
}

// TestScalarStructs round-trips the fixed-layout scalar structures.
func TestScalarStructs(t *testing.T) {
	roundTrip(t, "FAX_TIME", FAX_TIME{Hour: 22, Minute: 30})
	roundTrip(t, "FAX_VERSION", FAX_VERSION{
		DwSizeOfStruct:    28,
		BValid:            1,
		WMajorVersion:     10,
		WMinorVersion:     0,
		WMajorBuildNumber: 19041,
		WMinorBuildNumber: 0,
		DwFlags:           0,
	})
	roundTrip(t, "SYSTEMTIME", SYSTEMTIME{
		WYear: 2024, WMonth: 4, WDayOfWeek: 2, WDay: 23,
		WHour: 17, WMinute: 55, WSecond: 0, WMilliseconds: 0,
	})
	roundTrip(t, "FAX_SERVER_ACTIVITY", FAX_SERVER_ACTIVITY{
		DwSizeOfStruct:              36,
		DwIncomingMessages:          1,
		DwRoutingMessages:           2,
		DwOutgoingMessages:          3,
		DwDelegatedOutgoingMessages: 0,
		DwQueuedMessages:            4,
		DwErrorEvents:               0,
		DwWarningEvents:             1,
		DwInformationEvents:         5,
	})
	roundTrip(t, "FAX_MESSAGE_PROPS", FAX_MESSAGE_PROPS{DwValidityMask: 1, DwMsgFlags: 4})
	roundTrip(t, "FAX_OUTBOX_CONFIG", FAX_OUTBOX_CONFIG{
		DwSizeOfStruct: 44, BAllowPersonalCP: 1, BUseDeviceTSID: 0,
		DwRetries: 3, DwRetryDelay: 10,
		DtDiscountStart: FAX_TIME{Hour: 1}, DtDiscountEnd: FAX_TIME{Hour: 5},
		DwAgeLimit: 0, BBranding: 1,
	})
}

// TestStringStructs round-trips structures carrying [unique] wide-string pointers, both
// present and absent (nil), plus nested value structs and a SYSTEMTIME.
func TestStringStructs(t *testing.T) {
	roundTrip(t, "FAX_CONFIGURATIONW/full", FAX_CONFIGURATIONW{
		SizeOfStruct:     64,
		Retries:          3,
		RetryDelay:       10,
		DirtyDays:        30,
		Branding:         1,
		StartCheapTime:   FAX_TIME{Hour: 0, Minute: 0},
		StopCheapTime:    FAX_TIME{Hour: 8, Minute: 0},
		ArchiveDirectory: wstr(`C:\FaxArchive`),
		ProfileName:      wstr("Default"),
	})
	roundTrip(t, "FAX_CONFIGURATIONW/nil-strings", FAX_CONFIGURATIONW{SizeOfStruct: 64})
	roundTrip(t, "FAX_COVERPAGE_INFO_EXW", FAX_COVERPAGE_INFO_EXW{
		DwSizeOfStruct:          24,
		DwCoverPageFormat:       1,
		LpwstrCoverPageFileName: wstr("generic.cov"),
		BServerBased:            0,
		LpwstrNote:              wstr("note"),
		LpwstrSubject:           wstr("subject"),
	})
	roundTrip(t, "FAX_JOB_PARAMW", FAX_JOB_PARAMW{
		SizeOfStruct:    128,
		RecipientNumber: wstr("+15551234567"),
		RecipientName:   wstr("Alice"),
		SenderName:      wstr("Bob"),
		ScheduleAction:  0,
		ScheduleTime:    SYSTEMTIME{WYear: 2024, WMonth: 4, WDay: 23},
		CallHandle:      0,
		Reserved:        [3]uint64{},
	})
}

// TestConformantArrayStruct round-trips the [unique, size_is(field)] conformant-array
// shape (RPC_FAX_OUTBOUND_ROUTING_GROUPW.lpdwDevices), with a populated and an absent array.
func TestConformantArrayStruct(t *testing.T) {
	roundTrip(t, "RPC_FAX_OUTBOUND_ROUTING_GROUPW/devices", RPC_FAX_OUTBOUND_ROUTING_GROUPW{
		DwSizeOfStruct:  24,
		LpwstrGroupName: wstr("Group A"),
		DwNumDevices:    3,
		LpdwDevices:     []ndr.DWORD{10, 20, 30},
		Status:          FAX_GROUP_STATUS_ALL_DEV_VALID,
	})
	roundTrip(t, "RPC_FAX_OUTBOUND_ROUTING_GROUPW/empty", RPC_FAX_OUTBOUND_ROUTING_GROUPW{
		DwSizeOfStruct:  24,
		LpwstrGroupName: wstr("Group B"),
		DwNumDevices:    0,
		LpdwDevices:     nil,
		Status:          FAX_GROUP_STATUS_EMPTY,
	})
}

// TestRuleDestinationUnion exercises both arms of the FAX_RULE_DESTINATION switch union
// ([MS-FAX] 2.2.44): case 0 selects the device id, the default arm selects the group name.
func TestRuleDestinationUnion(t *testing.T) {
	roundTrip(t, "FAX_RULE_DESTINATION/device", FAX_RULE_DESTINATION{
		Tag:        0,
		DwDeviceId: 7,
	})
	roundTrip(t, "FAX_RULE_DESTINATION/group", FAX_RULE_DESTINATION{
		Tag:             1,
		LpwstrGroupName: ndr.WSTR("Group A"),
	})
	roundTrip(t, "RPC_FAX_OUTBOUND_ROUTING_RULEW", RPC_FAX_OUTBOUND_ROUTING_RULEW{
		DwSizeOfStruct:    24,
		DwAreaCode:        425,
		DwCountryCode:     1,
		LpwstrCountryName: wstr("United States"),
		Destination:       FAX_RULE_DESTINATION{Tag: 0, DwDeviceId: 3},
		BUseGroup:         0,
	})
}

// TestContextHandle confirms the RPC context handle is a fixed 20-byte value
// ([MS-RPCE] 2.3.2.2), the wire form for the FAX context handles.
func TestContextHandle(t *testing.T) {
	var h RPC_FAX_HANDLE
	for i := range h {
		h[i] = byte(i)
	}
	raw, err := ndr.Marshal(&struct{ H RPC_FAX_HANDLE }{H: h})
	if err != nil {
		t.Fatalf("Marshal RPC_FAX_HANDLE: %v", err)
	}
	if len(raw) != 20 {
		t.Errorf("RPC_FAX_HANDLE marshalled to %d bytes, want 20", len(raw))
	}
}
