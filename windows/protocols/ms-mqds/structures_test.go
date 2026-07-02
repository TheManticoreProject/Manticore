package msmqds

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// roundTrip marshals v under both transfer syntaxes and asserts it survives a round trip.
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

func propVarUI4(v uint32) msmqmq.PROPVARIANT {
	p := msmqmq.PROPVARIANT{Value: msmqmq.PropVariantUnion{UlVal: v}}
	p.SetVt(msmqmq.VT_UI4)
	return p
}

// TestMQSortKeyRoundTrip exercises the two-scalar sort key.
func TestMQSortKeyRoundTrip(t *testing.T) {
	roundTrip(t, "MQSORTKEY", MQSORTKEY{PropColumn: 42, DwOrder: QUERY_SORTDESCEND})
}

// TestMQSortSetRoundTrip exercises the [unique] pointer to a [size_is(cCol)] array of keys.
func TestMQSortSetRoundTrip(t *testing.T) {
	roundTrip(t, "MQSORTSET", MQSORTSET{
		CCol: 2,
		ACol: []MQSORTKEY{
			{PropColumn: 1, DwOrder: QUERY_SORTASCEND},
			{PropColumn: 2, DwOrder: QUERY_SORTDESCEND},
		},
	})
}

// TestMQColumnSetRoundTrip exercises the [unique] pointer to a [size_is(cCol)] PROPID array.
func TestMQColumnSetRoundTrip(t *testing.T) {
	roundTrip(t, "MQCOLUMNSET", MQCOLUMNSET{CCol: 3, ACol: []msmqmq.PROPID{10, 20, 30}})
}

// TestMQPropertyRestrictionRoundTrip exercises a restriction carrying a PROPVARIANT value.
func TestMQPropertyRestrictionRoundTrip(t *testing.T) {
	roundTrip(t, "MQPROPERTYRESTRICTION", MQPROPERTYRESTRICTION{
		Rel:   PREQ,
		Prop:  100,
		Prval: propVarUI4(0xDEADBEEF),
	})
}

// TestMQRestrictionRoundTrip exercises the [unique] pointer to a [size_is(cRes)] array of
// property restrictions, each carrying its own PROPVARIANT.
func TestMQRestrictionRoundTrip(t *testing.T) {
	roundTrip(t, "MQRESTRICTION", MQRESTRICTION{
		CRes: 2,
		PaPropRes: []MQPROPERTYRESTRICTION{
			{Rel: PRGE, Prop: 1, Prval: propVarUI4(7)},
			{Rel: PRLT, Prop: 2, Prval: propVarUI4(9)},
		},
	})
}

// TestMQRestrictionEmpty exercises the null [unique] array pointer (cRes == 0).
func TestMQRestrictionEmpty(t *testing.T) {
	roundTrip(t, "MQRESTRICTION-empty", MQRESTRICTION{CRes: 0, PaPropRes: nil})
}
