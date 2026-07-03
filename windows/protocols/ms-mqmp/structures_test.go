package msmqmp

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

// TestCACCreateRemoteCursorRoundTrip covers the plain three-DWORD cursor descriptor
// ([MS-MQMP] 2.2.4), which carries no pointers.
func TestCACCreateRemoteCursorRoundTrip(t *testing.T) {
	roundTrip(t, "CACCreateRemoteCursor", CACCreateRemoteCursor{
		HCursor:      0x11111111,
		Srv_hACQueue: 0x22222222,
		Cli_pQMQueue: 0x33333333,
	})
}

// TestObjectFormatArms covers OBJECT_FORMAT ([MS-MQMP] 2.2.3): the queue arm (ObjType 1,
// a unique pointer to a QUEUE_FORMAT) and the arm-less ObjType 2. SetObjType keeps the
// outer and mirrored discriminants consistent.
func TestObjectFormatArms(t *testing.T) {
	var qf msmqmq.QUEUE_FORMAT
	qf.SetQft(msmqmq.QUEUE_FORMAT_TYPE_PRIVATE)
	qf.Value.MOPrivateID = msmqmq.OBJECTID{Uniquifier: 7}

	var o OBJECT_FORMAT
	o.Value.PQueueFormat = &qf
	roundTrip(t, "OBJECT_FORMAT/queue", *o.SetObjType(1))

	o = OBJECT_FORMAT{}
	roundTrip(t, "OBJECT_FORMAT/objtype2", *o.SetObjType(2))
}
