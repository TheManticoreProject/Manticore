package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

func wstr(s string) *ndr.WSTR { w := ndr.WSTR(s); return &w }
func astr(s string) *ndr.STR  { a := ndr.STR(s); return &a }

// TestSCActionTypeWidth pins the [v1_enum] width: SC_ACTION_TYPE must marshal as 4 bytes,
// not the 2 bytes of a default NDR enum. SC_ACTION is {SC_ACTION_TYPE; DWORD} = 8 bytes.
func TestSCActionTypeWidth(t *testing.T) {
	raw, err := ndr.Marshal(&SC_ACTION{Type: SC_ACTION_RESTART, Delay: 0x44332211})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 8 {
		t.Fatalf("SC_ACTION marshalled to %d bytes, want 8 (v1_enum is 4 bytes): % x", len(raw), raw)
	}
	want := []byte{0x01, 0x00, 0x00, 0x00, 0x11, 0x22, 0x33, 0x44}
	if !reflect.DeepEqual(raw, want) {
		t.Errorf("SC_ACTION wire = % x, want % x", raw, want)
	}
}

// TestServiceStatusRoundTrip exercises the all-scalar SERVICE_STATUS record.
func TestServiceStatusRoundTrip(t *testing.T) {
	in := SERVICE_STATUS{DwServiceType: 0x10, DwCurrentState: 4, DwControlsAccepted: 1, DwWin32ExitCode: 0, DwCheckPoint: 7, DwWaitHint: 3000}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out SERVICE_STATUS
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("SERVICE_STATUS round-trip: got %+v want %+v", out, in)
	}
}

// TestQueryServiceConfigWRoundTrip exercises a struct of [unique] wide-string pointers,
// including a nil arm (DwTagId has no string after it).
func TestQueryServiceConfigWRoundTrip(t *testing.T) {
	in := QUERY_SERVICE_CONFIGW{
		DwServiceType:    0x10,
		DwStartType:      2,
		DwErrorControl:   1,
		LpBinaryPathName: wstr(`C:\Windows\System32\svc.exe`),
		LpDependencies:   nil,
		LpDisplayName:    wstr("My Service"),
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out QUERY_SERVICE_CONFIGW
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.LpBinaryPathName == nil || *out.LpBinaryPathName != *in.LpBinaryPathName {
		t.Errorf("LpBinaryPathName round-trip: %v", out.LpBinaryPathName)
	}
	if out.LpDependencies != nil {
		t.Errorf("LpDependencies should stay nil, got %v", out.LpDependencies)
	}
	if out.LpDisplayName == nil || *out.LpDisplayName != *in.LpDisplayName {
		t.Errorf("LpDisplayName round-trip: %v", out.LpDisplayName)
	}
}

// TestQueryServiceConfigARoundTrip is the ANSI counterpart, exercising *ndr.STR arms.
func TestQueryServiceConfigARoundTrip(t *testing.T) {
	in := QUERY_SERVICE_CONFIGA{
		DwServiceType:      0x10,
		LpBinaryPathName:   astr(`C:\svc.exe`),
		LpServiceStartName: astr("LocalSystem"),
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out QUERY_SERVICE_CONFIGA
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.LpBinaryPathName == nil || *out.LpBinaryPathName != *in.LpBinaryPathName {
		t.Errorf("LpBinaryPathName round-trip: %v", out.LpBinaryPathName)
	}
	if out.LpServiceStartName == nil || *out.LpServiceStartName != *in.LpServiceStartName {
		t.Errorf("LpServiceStartName round-trip: %v", out.LpServiceStartName)
	}
}

// TestSCRPCConfigInfoWUnion exercises the [switch_is] union with a [unique]-pointer arm
// (case 1 = SERVICE_DESCRIPTIONW). The discriminant is transmitted inline as a 4-byte tag.
func TestSCRPCConfigInfoWUnion(t *testing.T) {
	in := SC_RPC_CONFIG_INFOW{
		DwInfoLevel: 1,
		Field: SC_RPC_CONFIG_INFOW_Field{
			Tag: 1,
			Psd: &SERVICE_DESCRIPTIONW{LpDescription: wstr("a helpful service")},
		},
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out SC_RPC_CONFIG_INFOW
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.DwInfoLevel != 1 || out.Field.Tag != 1 {
		t.Fatalf("discriminant round-trip: DwInfoLevel=%d Tag=%d", out.DwInfoLevel, out.Field.Tag)
	}
	if out.Field.Psd == nil || out.Field.Psd.LpDescription == nil ||
		*out.Field.Psd.LpDescription != *in.Field.Psd.LpDescription {
		t.Errorf("union arm round-trip: %+v", out.Field.Psd)
	}
	if out.Field.Psfa != nil || out.Field.Psti != nil {
		t.Errorf("non-selected arms should be nil")
	}
}

// TestServiceTriggerRoundTrip exercises a [unique] GUID pointer plus a [unique] pointer to
// a conformant array of pointer-bearing structs (the data items).
func TestServiceTriggerRoundTrip(t *testing.T) {
	in := SERVICE_TRIGGER{
		DwTriggerType: 1,
		DwAction:      1,
		PDataItems: []SERVICE_TRIGGER_SPECIFIC_DATA_ITEM{
			{DwDataType: 2, PData: []uint8{1, 2, 3}},
			{DwDataType: 2, PData: []uint8{9}},
		},
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out SERVICE_TRIGGER
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.CDataItems != 2 || len(out.PDataItems) != 2 {
		t.Fatalf("CDataItems=%d len=%d, want 2/2", out.CDataItems, len(out.PDataItems))
	}
	if out.PDataItems[0].CbData != 3 || !reflect.DeepEqual(out.PDataItems[0].PData, []uint8{1, 2, 3}) {
		t.Errorf("data item 0 round-trip: %+v", out.PDataItems[0])
	}
}

// TestContextHandleSize pins the 20-byte context-handle representation ([MS-RPCE] 2.3.2.2).
func TestContextHandleSize(t *testing.T) {
	var h SC_RPC_HANDLE
	if !h.IsZero() {
		t.Fatal("zero-value handle should be zero")
	}
	raw, err := ndr.Marshal(&struct{ H SC_RPC_HANDLE }{})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 20 {
		t.Errorf("SC_RPC_HANDLE marshalled to %d bytes, want 20", len(raw))
	}
}
