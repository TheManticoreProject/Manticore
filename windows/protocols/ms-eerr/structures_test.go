package mseerr

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals v, unmarshals into a fresh value of the same type, and returns it
// for comparison. v and out must be pointers to the same struct type.
func roundTrip(t *testing.T, v any, out any) {
	t.Helper()
	raw, err := ndr.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%T): %v", v, err)
	}
	if err := ndr.Unmarshal(raw, out); err != nil {
		t.Fatalf("Unmarshal(%T): %v", v, err)
	}
}

// TestEEAString_RoundTrip exercises EEAString ([MS-EERR] 2.2.1.1): a [unique] pointer to
// a conformant byte array bounded by NLength, including the empty (nil pointer) case.
func TestEEAString_RoundTrip(t *testing.T) {
	cases := []EEAString{
		{NLength: 5, PString: []uint8("hello")},
		{NLength: 1, PString: []uint8{0x00}},
		{NLength: 0, PString: nil},
	}
	for _, in := range cases {
		var out EEAString
		roundTrip(t, &in, &out)
		if !reflect.DeepEqual(in, out) {
			t.Fatalf("EEAString round-trip: got %+v want %+v", out, in)
		}
	}
}

// TestEEUString_RoundTrip exercises EEUString ([MS-EERR] 2.2.1.2): a [unique] pointer to
// a conformant UTF-16 array bounded by NLength (a character count).
func TestEEUString_RoundTrip(t *testing.T) {
	cases := []EEUString{
		{NLength: 4, PString: []uint16{'H', 'O', 'S', 'T'}},
		{NLength: 0, PString: nil},
	}
	for _, in := range cases {
		var out EEUString
		roundTrip(t, &in, &out)
		if !reflect.DeepEqual(in, out) {
			t.Fatalf("EEUString round-trip: got %+v want %+v", out, in)
		}
	}
}

// TestBinaryEEInfo_RoundTrip exercises BinaryEEInfo ([MS-EERR] 2.2.1.3): a [unique]
// pointer to a conformant byte blob bounded by NSize.
func TestBinaryEEInfo_RoundTrip(t *testing.T) {
	cases := []BinaryEEInfo{
		{NSize: 3, PBlob: []uint8{0xde, 0xad, 0xbe}},
		{NSize: 0, PBlob: nil},
	}
	for _, in := range cases {
		var out BinaryEEInfo
		roundTrip(t, &in, &out)
		if !reflect.DeepEqual(in, out) {
			t.Fatalf("BinaryEEInfo round-trip: got %+v want %+v", out, in)
		}
	}
}

// TestExtendedErrorParam_RoundTrip exercises every arm of the non-encapsulated union
// ExtendedErrorParam ([MS-EERR] 2.2.2), including the empty eeptiNone arm. The union's
// switch_type(short) discriminant (Field.Tag) must match the case value.
func TestExtendedErrorParam_RoundTrip(t *testing.T) {
	cases := []ExtendedErrorParam{
		{Type: EeptiAnsiString, Field: ExtendedErrorParam_Field{Tag: 1, AnsiString: EEAString{NLength: 3, PString: []uint8("abc")}}},
		{Type: EeptiUnicodeString, Field: ExtendedErrorParam_Field{Tag: 2, UnicodeString: EEUString{NLength: 2, PString: []uint16{'o', 'k'}}}},
		{Type: EeptiLongVal, Field: ExtendedErrorParam_Field{Tag: 3, LVal: -12345}},
		{Type: EeptiShortValue, Field: ExtendedErrorParam_Field{Tag: 4, IVal: 777}},
		{Type: EeptiPointerValue, Field: ExtendedErrorParam_Field{Tag: 5, PVal: 0x0011223344556677}},
		{Type: EeptiNone, Field: ExtendedErrorParam_Field{Tag: 6}},
		{Type: EeptiBinary, Field: ExtendedErrorParam_Field{Tag: 7, Blob: BinaryEEInfo{NSize: 2, PBlob: []uint8{0x01, 0x02}}}},
	}
	for _, in := range cases {
		var out ExtendedErrorParam
		roundTrip(t, &in, &out)
		if !reflect.DeepEqual(in, out) {
			t.Fatalf("ExtendedErrorParam(Type=%d) round-trip: got %+v want %+v", in.Type, out, in)
		}
	}
}

// TestEEComputerName_RoundTrip exercises both arms of EEComputerName ([MS-EERR] 2.2.3):
// eecnpPresent (carries an EEUString name) and eecnpNotPresent (empty arm).
func TestEEComputerName_RoundTrip(t *testing.T) {
	cases := []EEComputerName{
		{Type: EecnpPresent, Field: EEComputerName_Field{Tag: 1, Name: EEUString{NLength: 3, PString: []uint16{'D', 'C', '1'}}}},
		{Type: EecnpNotPresent, Field: EEComputerName_Field{Tag: 2}},
	}
	for _, in := range cases {
		var out EEComputerName
		roundTrip(t, &in, &out)
		if !reflect.DeepEqual(in, out) {
			t.Fatalf("EEComputerName(Type=%d) round-trip: got %+v want %+v", in.Type, out, in)
		}
	}
}

// TestExtendedErrorInfo_RoundTrip exercises a single ExtendedErrorInfo node ([MS-EERR]
// 2.2.4) with Next nil: the embedded conformant Params array (max_count hoisted to the
// head of the struct) and the scalar members.
func TestExtendedErrorInfo_RoundTrip(t *testing.T) {
	in := ExtendedErrorInfo{
		Next:                nil,
		ComputerName:        EEComputerName{Type: EecnpPresent, Field: EEComputerName_Field{Tag: 1, Name: EEUString{NLength: 2, PString: []uint16{'X', 'Y'}}}},
		ProcessID:           4242,
		TimeStamp:           0x01D9AABBCCDDEEFF,
		GeneratingComponent: 7,
		Status:              0xC0000022,
		DetectionLocation:   123,
		Flags:               0,
		NLen:                2,
		Params: []ExtendedErrorParam{
			{Type: EeptiLongVal, Field: ExtendedErrorParam_Field{Tag: 3, LVal: 1}},
			{Type: EeptiBinary, Field: ExtendedErrorParam_Field{Tag: 7, Blob: BinaryEEInfo{NSize: 1, PBlob: []uint8{0xff}}}},
		},
	}
	var out ExtendedErrorInfo
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("ExtendedErrorInfo round-trip: got %+v want %+v", out, in)
	}
}

// TestExtendedErrorInfo_Chain exercises the [unique] self-referential Next pointer
// ([MS-EERR] 2.2.4): a two-node error chain terminated by a nil Next.
func TestExtendedErrorInfo_Chain(t *testing.T) {
	in := ExtendedErrorInfo{
		ComputerName:      EEComputerName{Type: EecnpNotPresent, Field: EEComputerName_Field{Tag: 2}},
		ProcessID:         1,
		Status:            0x00000103,
		DetectionLocation: 1,
		NLen:              1,
		Params:            []ExtendedErrorParam{{Type: EeptiShortValue, Field: ExtendedErrorParam_Field{Tag: 4, IVal: 9}}},
		Next: &ExtendedErrorInfo{
			ComputerName:      EEComputerName{Type: EecnpNotPresent, Field: EEComputerName_Field{Tag: 2}},
			ProcessID:         2,
			Status:            0x00000000,
			DetectionLocation: 2,
			NLen:              1,
			Params:            []ExtendedErrorParam{{Type: EeptiLongVal, Field: ExtendedErrorParam_Field{Tag: 3, LVal: -1}}},
			Next:              nil,
		},
	}
	var out ExtendedErrorInfo
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("ExtendedErrorInfo chain round-trip: got %+v want %+v", out, in)
	}
}
