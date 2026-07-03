package msmqqp

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
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

// TestREMOTEREADACKValues pins the NDR enum values ([MS-MQQP] 2.2.2.2). The type is a
// 16-bit NDR enum; the eAckNack field is reserved on the wire.
func TestREMOTEREADACKValues(t *testing.T) {
	if RR_UNKNOWN != 0 || RR_NACK != 1 || RR_ACK != 2 {
		t.Fatalf("REMOTEREADACK values = %d/%d/%d, want 0/1/2", RR_UNKNOWN, RR_NACK, RR_ACK)
	}
}

// TestREMOTEREADDESCEmpty covers the client-side shape of REMOTEREADDESC ([MS-MQQP] 2.2.2.1):
// dwSize is 0 and the unique lpBuffer pointer is NULL, as the client MUST send it.
func TestREMOTEREADDESCEmpty(t *testing.T) {
	roundTrip(t, "REMOTEREADDESC/empty", REMOTEREADDESC{
		HRemoteQueue: 0x11111111,
		HCursor:      0x22222222,
		UlAction:     0x00000000, // MQ_ACTION_RECEIVE
		UlTimeout:    0xFFFFFFFF, // infinite
		DwSize:       0,
		DwQueue:      0x11111111,
		DwRequestID:  0x0000002A,
		EAckNack:     RR_UNKNOWN,
	})
}

// TestREMOTEREADDESCBuffer exercises the conformant varying byte array carried by the unique
// lpBuffer pointer (size_is/length_is = dwSize), the server-side shape of the structure.
func TestREMOTEREADDESCBuffer(t *testing.T) {
	buf := []uint8{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02, 0x03}
	roundTrip(t, "REMOTEREADDESC/buffer", REMOTEREADDESC{
		HRemoteQueue: 0x11111111,
		UlAction:     0x80000000, // MQ_ACTION_PEEK_CURRENT
		DwSize:       uint32(len(buf)),
		DwArriveTime: 0x5F000000,
		EAckNack:     RR_ACK,
		LpBuffer:     buf,
	})
}

// TestREMOTEREADDESC2 covers the v2 descriptor ([MS-MQQP] 2.2.2.3): a unique pointer to a
// REMOTEREADDESC plus a sequential id. Both the populated and NULL-pointer arms are tested.
func TestREMOTEREADDESC2(t *testing.T) {
	buf := []uint8{0x11, 0x22, 0x33, 0x44}
	roundTrip(t, "REMOTEREADDESC2/populated", REMOTEREADDESC2{
		PRemoteReadDesc: &REMOTEREADDESC{
			HRemoteQueue: 0xAABBCCDD,
			DwSize:       uint32(len(buf)),
			LpBuffer:     buf,
		},
		SequentialId: 0x0102030405060708,
	})

	roundTrip(t, "REMOTEREADDESC2/null", REMOTEREADDESC2{
		PRemoteReadDesc: nil,
		SequentialId:    0xCAFEF00DDEADBEEF,
	})
}
