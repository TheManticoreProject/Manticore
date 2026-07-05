package mstsch

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

// roundTripTsch marshals in, unmarshals into a fresh value of the same type, and asserts
// deep equality. Named distinctly to avoid colliding with helpers in the same package.
func roundTripTsch[T any](t *testing.T, name string, in T) {
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

// TestAT_ENUM_RoundTrip exercises AT_ENUM, which carries a [unique] wide-string Command.
func TestAT_ENUM_RoundTrip(t *testing.T) {
	roundTripTsch(t, "AT_ENUM", AT_ENUM{
		JobId:       7,
		JobTime:     3600000,
		DaysOfMonth: 0x00000005,
		DaysOfWeek:  0x02,
		Flags:       0x11,
		Command:     wstr("C:\\Windows\\System32\\cmd.exe /c echo hi"),
	})
	roundTripTsch(t, "AT_ENUM/nil-command", AT_ENUM{JobId: 1})
}

// TestAT_INFO_RoundTrip exercises AT_INFO.
func TestAT_INFO_RoundTrip(t *testing.T) {
	roundTripTsch(t, "AT_INFO", AT_INFO{
		JobTime:     43200000,
		DaysOfMonth: 0,
		DaysOfWeek:  0x7F,
		Flags:       0x08,
		Command:     wstr("notepad.exe"),
	})
}

// TestAT_ENUM_CONTAINER_RoundTrip exercises a [unique] pointer to a conformant array of
// AT_ENUM structs that each carry a [unique] string pointer.
func TestAT_ENUM_CONTAINER_RoundTrip(t *testing.T) {
	roundTripTsch(t, "AT_ENUM_CONTAINER", AT_ENUM_CONTAINER{
		EntriesRead: 2,
		Buffer: []AT_ENUM{
			{JobId: 1, JobTime: 1000, DaysOfWeek: 0x01, Command: wstr("a.exe")},
			{JobId: 2, JobTime: 2000, DaysOfWeek: 0x40, Command: wstr("b.exe")},
		},
	})
	roundTripTsch(t, "AT_ENUM_CONTAINER/empty", AT_ENUM_CONTAINER{EntriesRead: 0, Buffer: []AT_ENUM{}})
}

// TestTASK_USER_CRED_RoundTrip exercises TASK_USER_CRED with its two [unique] strings.
func TestTASK_USER_CRED_RoundTrip(t *testing.T) {
	roundTripTsch(t, "TASK_USER_CRED", TASK_USER_CRED{
		UserId:   wstr("DOMAIN\\user"),
		Password: wstr("s3cr3t"),
		Flags:    uint32(CredFlagDefault),
	})
	roundTripTsch(t, "TASK_USER_CRED/nil", TASK_USER_CRED{Flags: 0})
}

// TestTASK_XML_ERROR_INFO_RoundTrip exercises TASK_XML_ERROR_INFO.
func TestTASK_XML_ERROR_INFO_RoundTrip(t *testing.T) {
	roundTripTsch(t, "TASK_XML_ERROR_INFO", TASK_XML_ERROR_INFO{
		Line:   12,
		Column: 34,
		Node:   wstr("Triggers"),
		Value:  wstr("bad value"),
	})
}
