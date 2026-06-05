package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTripTransportMisc marshals in, unmarshals into a fresh value of the same
// type, and asserts the result is deeply equal to in. It is named distinctly to
// avoid colliding with helpers defined by other test files in this package.
func roundTripTransportMisc[T any](t *testing.T, name string, in T) {
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

// TestSERVER_TRANSPORT_INFO_2_RoundTrip exercises a struct that carries a
// [unique] pointer to a conformant byte array (size_is) alongside several
// [unique] string pointers.
func TestSERVER_TRANSPORT_INFO_2_RoundTrip(t *testing.T) {
	in := SERVER_TRANSPORT_INFO_2{
		Svti2Numberofvcs:            7,
		Svti2Transportname:          "\\Device\\NetbiosSmb",
		Svti2Transportaddress:       []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		Svti2Transportaddresslength: 5,
		Svti2Networkaddress:         "HOST",
		Svti2Domain:                 "WORKGROUP",
		Svti2Flags:                  0x00000001,
	}
	roundTripTransportMisc(t, "SERVER_TRANSPORT_INFO_2", in)
}

// TestTIME_OF_DAY_INFO_RoundTrip exercises a flat all-scalar struct including
// the signed TodTimezone field.
func TestTIME_OF_DAY_INFO_RoundTrip(t *testing.T) {
	in := TIME_OF_DAY_INFO{
		TodElapsedt:  1700000000,
		TodMsecs:     123,
		TodHours:     14,
		TodMins:      30,
		TodSecs:      45,
		TodHunds:     50,
		TodTimezone:  -120,
		TodTinterval: 310,
		TodDay:       4,
		TodMonth:     6,
		TodYear:      2026,
		TodWeekday:   4,
	}
	roundTripTransportMisc(t, "TIME_OF_DAY_INFO", in)
}
