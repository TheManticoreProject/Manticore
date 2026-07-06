package mspan

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts the
// result is deeply equal to in. This is the wire-shape acceptance gate for the MS-PAN
// NDR types in the absence of a live Print System Asynchronous Notification server.
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

// TestContextHandles exercises the two 20-octet context handles. NDR marshals top-level
// structs only, so each is wrapped as it appears on the wire (an inline field).
func TestContextHandles(t *testing.T) {
	type remoteObj struct{ H PRPCREMOTEOBJECT }
	type notifyObj struct{ H PNOTIFYOBJECT }

	roundTrip(t, "PRPCREMOTEOBJECT", remoteObj{H: PRPCREMOTEOBJECT{
		0x01, 0x00, 0x00, 0x00, 0xde, 0xad, 0xbe, 0xef, 0x00, 0x11,
		0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb,
	}})
	roundTrip(t, "PNOTIFYOBJECT", notifyObj{H: PNOTIFYOBJECT{
		0x00, 0x00, 0x00, 0x00, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab,
		0xcd, 0xef, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
	}})
}

// TestPrintAsyncNotificationType exercises the GUID typedef and locks its 16-octet wire
// size (Data1/Data2/Data3 + Data4[8]), the value the notification-type identifier is
// carried as.
func TestPrintAsyncNotificationType(t *testing.T) {
	type wrap struct{ T PrintAsyncNotificationType }
	in := wrap{T: PrintAsyncNotificationType(msdtyp.GUID{
		Data1: 0x23cbe492,
		Data2: 0xf6bf,
		Data3: 0x4b5b,
		Data4: [8]byte{0xa0, 0x66, 0x8e, 0x2b, 0x1e, 0x3b, 0x1d, 0x2c},
	})}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 16 {
		t.Fatalf("PrintAsyncNotificationType wire size = %d, want 16 (typedef GUID)", len(raw))
	}
	roundTrip(t, "PrintAsyncNotificationType", in)
}

// TestEnumsAreV1Enum locks the [v1_enum] 32-bit width of both MS-PAN enums. A 16-bit
// default enum would silently emit 2 octets and corrupt the wire.
func TestEnumsAreV1Enum(t *testing.T) {
	type conv struct {
		V PrintAsyncNotifyConversationStyle
	}
	type filt struct{ V PrintAsyncNotifyUserFilter }

	raw, err := ndr.Marshal(&conv{V: KUniDirectional})
	if err != nil {
		t.Fatalf("Marshal conversation style: %v", err)
	}
	if len(raw) != 4 {
		t.Fatalf("PrintAsyncNotifyConversationStyle wire size = %d, want 4 (v1_enum)", len(raw))
	}
	raw, err = ndr.Marshal(&filt{V: KAllUsers})
	if err != nil {
		t.Fatalf("Marshal user filter: %v", err)
	}
	if len(raw) != 4 {
		t.Fatalf("PrintAsyncNotifyUserFilter wire size = %d, want 4 (v1_enum)", len(raw))
	}

	roundTrip(t, "ConversationStyle/bidi", conv{V: KBiDirectional})
	roundTrip(t, "UserFilter/perUser", filt{V: KPerUser})
}

// TestChannelArray exercises the IRPCAsyncNotify_GetNewChannel [out] shape:
// [size_is( , *pNoOfChannels)] PNOTIFYOBJECT** — a [unique] pointer to a conformant array
// of context handles, sized by a sibling count field.
func TestChannelArray(t *testing.T) {
	type resp struct {
		Count    ndr.DWORD
		Channels []PNOTIFYOBJECT `ndr:"unique,size_is=Count"`
	}
	roundTrip(t, "channels/populated", resp{
		Count: 2,
		Channels: []PNOTIFYOBJECT{
			{0x01, 0x02, 0x03, 0x04},
			{0x05, 0x06, 0x07, 0x08},
		},
	})
	roundTrip(t, "channels/nil", resp{Count: 0, Channels: nil})
}

// TestNotificationBlob exercises the [size_is(N), unique] byte* notification-data shape
// used by GetNotification / GetNotificationSendResponse / CloseChannel.
func TestNotificationBlob(t *testing.T) {
	type msg struct {
		Size ndr.DWORD
		Data []byte `ndr:"unique,size_is=Size"`
	}
	roundTrip(t, "blob/populated", msg{Size: 5, Data: []byte{0xde, 0xad, 0xbe, 0xef, 0x2a}})
	roundTrip(t, "blob/nil", msg{Size: 0, Data: nil})
}

// TestOptionalString exercises the [string, unique] wchar_t* shape used by pName and
// ppRmtServerReferral: a [unique] pointer that may be nil.
func TestOptionalString(t *testing.T) {
	type req struct {
		Name *ndr.WSTR `ndr:"unique"`
	}
	name := ndr.WSTR(`\\server\printer`)
	roundTrip(t, "name/present", req{Name: &name})
	roundTrip(t, "name/nil", req{Name: nil})
}
