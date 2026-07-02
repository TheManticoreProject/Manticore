package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts the
// result is deeply equal to in. This surfaces wire-shape (tag) bugs without a live server.
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

// TestFILETIME_RoundTrip / TestSYSTEMTIME_RoundTrip exercise the [MS-DTYP] fillers that the
// MS-RPRN IDL references but does not define.
func TestFILETIME_RoundTrip(t *testing.T) {
	roundTrip(t, "FILETIME", FILETIME{DwLowDateTime: 0xDEADBEEF, DwHighDateTime: 0x01D9C0DE})
}

func TestSYSTEMTIME_RoundTrip(t *testing.T) {
	roundTrip(t, "SYSTEMTIME", SYSTEMTIME{
		WYear: 2026, WMonth: 7, WDayOfWeek: 3, WDay: 1,
		WHour: 13, WMinute: 37, WSecond: 42, WMilliseconds: 500,
	})
}

// TestPRINTER_INFO_1_RoundTrip exercises a struct of [unique] wide-string pointers,
// including a nil pointer (NULL referent).
func TestPRINTER_INFO_1_RoundTrip(t *testing.T) {
	roundTrip(t, "PRINTER_INFO_1 full", PRINTER_INFO_1{
		Flags:        1,
		PDescription: wstr("HP LaserJet"),
		PName:        wstr("\\\\srv\\prn1"),
		PComment:     wstr("front desk"),
	})
	roundTrip(t, "PRINTER_INFO_1 nil comment", PRINTER_INFO_1{
		Flags: 2,
		PName: wstr("\\\\srv\\prn2"),
	})
}

// TestPRINTER_CONTAINER_RoundTrip exercises the classic MS-RPRN container: a Level field
// plus a [switch_is(Level)] union whose selected arm is a [unique] pointer to an info struct.
func TestPRINTER_CONTAINER_RoundTrip(t *testing.T) {
	in := PRINTER_CONTAINER{
		Level: 1,
		PrinterInfo: PRINTER_CONTAINER_PrinterInfo{
			Tag:           1,
			PPrinterInfo1: &PRINTER_INFO_1{Flags: 4, PName: wstr("prn")},
		},
	}
	roundTrip(t, "PRINTER_CONTAINER", in)
}

// TestRPC_PrintPropertyValue_RoundTrip exercises the hand-modeled RPC_PrintPropertyValue
// union across its arms: the wide-string arm, the scalar arms, and the counted-blob arm.
func TestRPC_PrintPropertyValue_RoundTrip(t *testing.T) {
	roundTrip(t, "prop string", RPC_PrintPropertyValue{
		EPropertyType: kRpcPropertyTypeString,
		Value:         RPC_PrintPropertyValue_Value{Tag: kRpcPropertyTypeString, PropertyString: wstr("value")},
	})
	roundTrip(t, "prop int32", RPC_PrintPropertyValue{
		EPropertyType: kRpcPropertyTypeInt32,
		Value:         RPC_PrintPropertyValue_Value{Tag: kRpcPropertyTypeInt32, PropertyInt32: -12345},
	})
	roundTrip(t, "prop int64", RPC_PrintPropertyValue{
		EPropertyType: kRpcPropertyTypeInt64,
		Value:         RPC_PrintPropertyValue_Value{Tag: kRpcPropertyTypeInt64, PropertyInt64: 0x0123456789ABCDEF},
	})
	roundTrip(t, "prop blob", RPC_PrintPropertyValue{
		EPropertyType: kRpcPropertyTypeBuffer,
		Value: RPC_PrintPropertyValue_Value{
			Tag:          kRpcPropertyTypeBuffer,
			PropertyBlob: RPC_PrintPropertyBlob{CbBuf: 4, PBuf: []uint8{0xDE, 0xAD, 0xBE, 0xEF}},
		},
	})
}

// TestRPC_V2_NOTIFY_INFO_DATA_DATA_RoundTrip exercises the union whose case values were the
// TABLE_* #define constants; a wrong (symbolic) case label would fail to marshal the arm.
func TestRPC_V2_NOTIFY_INFO_DATA_DATA_RoundTrip(t *testing.T) {
	roundTrip(t, "notify dword", RPC_V2_NOTIFY_INFO_DATA_DATA{Tag: 1, DwData: 0xCAFEF00D})
	roundTrip(t, "notify string", RPC_V2_NOTIFY_INFO_DATA_DATA{
		Tag:    2,
		String: STRING_CONTAINER{CbBuf: 6, PszString: []uint16{'h', 'i', 0}},
	})
	roundTrip(t, "notify time", RPC_V2_NOTIFY_INFO_DATA_DATA{
		Tag:        4,
		SystemTime: SYSTEMTIME_CONTAINER{CbBuf: 16, PSystemTime: &SYSTEMTIME{WYear: 2026, WMonth: 7}},
	})
}

// TestRPC_V2_NOTIFY_INFO_RoundTrip exercises a [unique] pointer to a conformant array of
// pointer-bearing structs (each element embeds the notify-data union).
func TestRPC_V2_NOTIFY_INFO_RoundTrip(t *testing.T) {
	in := RPC_V2_NOTIFY_INFO{
		Version: 2,
		Flags:   0,
		Count:   2,
		AData: []RPC_V2_NOTIFY_INFO_DATA{
			{Type: 1, Field: 2, Reserved: 1, Id: 10, Data: RPC_V2_NOTIFY_INFO_DATA_DATA{Tag: 1, DwData: 7}},
			{Type: 1, Field: 3, Reserved: 2, Id: 11, Data: RPC_V2_NOTIFY_INFO_DATA_DATA{Tag: 2, String: STRING_CONTAINER{CbBuf: 2, PszString: []uint16{'x'}}}},
		},
	}
	roundTrip(t, "RPC_V2_NOTIFY_INFO", in)
}
