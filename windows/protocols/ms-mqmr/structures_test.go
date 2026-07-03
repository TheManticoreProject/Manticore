package msmqmr

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// roundTrip marshals v under both transfer syntaxes, unmarshals into a fresh value of the
// same type, and asserts the result equals v.
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	for _, s := range []ndr.Syntax{ndr.NDR20, ndr.NDR64} {
		raw, err := ndr.MarshalAs(&in, s)
		if err != nil {
			t.Fatalf("%s %s marshal: %v", name, s, err)
		}
		var out T
		if err := ndr.UnmarshalAs(raw, &out, s); err != nil {
			t.Fatalf("%s %s unmarshal: %v", name, s, err)
		}
		if !reflect.DeepEqual(in, out) {
			t.Errorf("%s %s round-trip:\n got %+v\nwant %+v", name, s, out, in)
		}
	}
}

// TestMgmtObjectArms exercises each selectable MGMT_OBJECT union arm. SetType keeps the
// outer discriminant and the union's mirrored discriminant consistent, matching the
// non-encapsulated union wire form ([C706] 14.3.8): the discriminant is transmitted twice.
func TestMgmtObjectArms(t *testing.T) {
	// MGMT_MACHINE: the reserved DWORD arm.
	var m MGMT_OBJECT
	m = MGMT_OBJECT{Value: MgmtObjectUnion{Reserved1: 0xCAFEF00D}}
	roundTrip(t, "MGMT_MACHINE", *m.SetType(MGMT_MACHINE))

	// MGMT_QUEUE: a unique pointer to a direct-format QUEUE_FORMAT ([MS-MQMQ] 2.2.7).
	qf := (&msmqmq.QUEUE_FORMAT{Value: msmqmq.QueueFormatUnion{MPDirectID: wstr(`OS:server\queue`)}}).
		SetQft(msmqmq.QUEUE_FORMAT_TYPE_DIRECT)
	m = MGMT_OBJECT{Value: MgmtObjectUnion{PQueueFormat: qf}}
	roundTrip(t, "MGMT_QUEUE", *m.SetType(MGMT_QUEUE))

	// MGMT_SESSION: the reserved DWORD arm.
	m = MGMT_OBJECT{Value: MgmtObjectUnion{Reserved2: 7}}
	roundTrip(t, "MGMT_SESSION", *m.SetType(MGMT_SESSION))
}

func wstr(s string) *ndr.WSTR { w := ndr.WSTR(s); return &w }
