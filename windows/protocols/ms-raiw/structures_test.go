package msraiw

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// syntaxes is the set of NDR transfer syntaxes every round-trip is exercised under.
var syntaxes = []ndr.Syntax{ndr.NDR20, ndr.NDR64}

// TestEnumWidths confirms every MS-RAIW enum marshals as a 16-bit NDR enum under the
// NDR20 transfer syntax ([C706] section 14.3.6): 2 octets, not 4.
func TestEnumWidths(t *testing.T) {
	type holder struct{ V WINSINTF_ACT_E }
	raw, err := ndr.Marshal(&holder{V: WINSINTF_E_QUERY})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if len(raw) != 2 {
		t.Fatalf("WINSINTF_ACT_E marshalled to %d bytes, want 2 (NDR enum width)", len(raw))
	}
	var out holder
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.V != WINSINTF_E_QUERY {
		t.Fatalf("enum round-trip: got %d want %d", out.V, WINSINTF_E_QUERY)
	}
}

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts
// the result is deeply equal to in — under both NDR transfer syntaxes.
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	for _, s := range syntaxes {
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

func TestWINSINTF_ADD_T_RoundTrip(t *testing.T) {
	roundTrip(t, "WINSINTF_ADD_T", WINSINTF_ADD_T{Type: 1, Len: 4, IPAdd: 0xC0A80001})
}

func TestWINSINTF_ADD_VERS_MAP_T_RoundTrip(t *testing.T) {
	roundTrip(t, "WINSINTF_ADD_VERS_MAP_T", WINSINTF_ADD_VERS_MAP_T{
		Add:    WINSINTF_ADD_T{Type: 1, Len: 4, IPAdd: 0x0A000001},
		VersNo: dtyp.LARGE_INTEGER(0x0000000100000005),
	})
}

func TestWINSINTF_SCV_REQ_T_RoundTrip(t *testing.T) {
	roundTrip(t, "WINSINTF_SCV_REQ_T", WINSINTF_SCV_REQ_T{
		Opcode_e: WINSINTF_E_SCV_VERIFY,
		Age:      3600,
		FForce:   1,
	})
}

// TestWINSINTF_RECORD_ACTION_T_RoundTrip exercises the record shape: a leading 16-bit
// enum, a [unique] conformant byte buffer (PName), a [unique,size_is] array of
// pointer-free structs (PAdd), an inline struct, a LARGE_INTEGER and the DWORD_PTR
// timestamp.
func TestWINSINTF_RECORD_ACTION_T_RoundTrip(t *testing.T) {
	name := []byte("WINSSRV\x00")
	roundTrip(t, "WINSINTF_RECORD_ACTION_T", WINSINTF_RECORD_ACTION_T{
		Cmd_e:      WINSINTF_E_INSERT,
		PName:      name,
		NameLen:    ndr.DWORD(len(name) - 1),
		TypOfRec_e: 0,
		NoOfAdds:   2,
		PAdd: []WINSINTF_ADD_T{
			{Type: 1, Len: 4, IPAdd: 0x0A000001},
			{Type: 1, Len: 4, IPAdd: 0x0A000002},
		},
		Add:       WINSINTF_ADD_T{Type: 1, Len: 4, IPAdd: 0x0A000003},
		VersNo:    dtyp.LARGE_INTEGER(0x000000020000000A),
		NodeTyp:   1,
		OwnerId:   7,
		State_e:   0,
		FStatic:   0,
		TimeStamp: 0xDEADBEEF,
	})
}

// TestWINSINTF_RECORD_ACTION_T_NilPointers confirms the [unique] pointers marshal as a
// null referent (and round-trip back to nil) when unset.
func TestWINSINTF_RECORD_ACTION_T_NilPointers(t *testing.T) {
	roundTrip(t, "WINSINTF_RECORD_ACTION_T(nil)", WINSINTF_RECORD_ACTION_T{
		Cmd_e:   WINSINTF_E_QUERY,
		NameLen: 0,
		Add:     WINSINTF_ADD_T{Type: 1, Len: 4, IPAdd: 0x7F000001},
		VersNo:  dtyp.LARGE_INTEGER(0),
	})
}

// TestWINSINTF_RECS_T_RoundTrip exercises the [unique] pointer to a conformant array of
// records, each carrying its own [unique] pointers.
func TestWINSINTF_RECS_T_RoundTrip(t *testing.T) {
	rec := WINSINTF_RECORD_ACTION_T{
		Cmd_e:    WINSINTF_E_INSERT,
		PName:    []byte("HOST\x00"),
		NameLen:  4,
		NoOfAdds: 1,
		PAdd:     []WINSINTF_ADD_T{{Type: 1, Len: 4, IPAdd: 0x0A000005}},
		Add:      WINSINTF_ADD_T{Type: 1, Len: 4, IPAdd: 0x0A000005},
		VersNo:   dtyp.LARGE_INTEGER(1),
	}
	roundTrip(t, "WINSINTF_RECS_T", WINSINTF_RECS_T{
		BuffSize:      1,
		PRow:          []WINSINTF_RECORD_ACTION_T{rec},
		NoOfRecs:      1,
		TotalNoOfRecs: 1,
	})
}

// statWithData builds a populated WINSINTF_STAT_T (nested Counters/TimeStamps structs,
// SYSTEMTIME members, and a [unique,size_is] replication-partner array).
func statWithData() WINSINTF_STAT_T {
	ts := dtyp.SYSTEMTIME{WYear: 2024, WMonth: 4, WDay: 23, WHour: 12}
	return WINSINTF_STAT_T{
		Counters: WINSINTF_STAT_T_Counters{
			NoOfUniqueReg: 10, NoOfGroupReg: 20, NoOfQueries: 30,
			NoOfSuccQueries: 25, NoOfFailQueries: 5, NoOfUniqueRef: 1,
			NoOfGroupRef: 2, NoOfRel: 3, NoOfSuccRel: 3, NoOfFailRel: 0,
			NoOfUniqueCnf: 1, NoOfGroupCnf: 1,
		},
		TimeStamps: WINSINTF_STAT_T_TimeStamps{
			WINSStartTime: ts, LastPScvTime: ts, CounterResetTime: ts,
		},
		NoOfPnrs: 1,
		PRplPnrs: []WINSINTF_RPL_COUNTERS_T{
			{Add: WINSINTF_ADD_T{Type: 1, Len: 4, IPAdd: 0x0A000009}, NoOfRpls: 4, NoOfCommFails: 0},
		},
	}
}

func TestWINSINTF_STAT_T_RoundTrip(t *testing.T) {
	roundTrip(t, "WINSINTF_STAT_T", statWithData())
}

// TestWINSINTF_RESULTS_T_RoundTrip exercises the fixed-size [25] AddVersMaps array plus
// an embedded WINSINTF_STAT_T (which itself defers a conformant array behind a pointer).
func TestWINSINTF_RESULTS_T_RoundTrip(t *testing.T) {
	var res WINSINTF_RESULTS_T
	res.NoOfOwners = 2
	res.AddVersMaps[0] = WINSINTF_ADD_VERS_MAP_T{
		Add: WINSINTF_ADD_T{Type: 1, Len: 4, IPAdd: 0x0A000001}, VersNo: dtyp.LARGE_INTEGER(5),
	}
	res.AddVersMaps[1] = WINSINTF_ADD_VERS_MAP_T{
		Add: WINSINTF_ADD_T{Type: 1, Len: 4, IPAdd: 0x0A000002}, VersNo: dtyp.LARGE_INTEGER(9),
	}
	res.MyMaxVersNo = dtyp.LARGE_INTEGER(9)
	res.RefreshInterval = 345600
	res.WINSStat = statWithData()
	roundTrip(t, "WINSINTF_RESULTS_T", res)
}

// TestWINSINTF_RESULTS_NEW_T_RoundTrip exercises the [unique,size_is] pointer variant of
// the owner-version map alongside the embedded statistics block.
func TestWINSINTF_RESULTS_NEW_T_RoundTrip(t *testing.T) {
	roundTrip(t, "WINSINTF_RESULTS_NEW_T", WINSINTF_RESULTS_NEW_T{
		NoOfOwners: 2,
		PAddVersMaps: []WINSINTF_ADD_VERS_MAP_T{
			{Add: WINSINTF_ADD_T{Type: 1, Len: 4, IPAdd: 0x0A000001}, VersNo: dtyp.LARGE_INTEGER(5)},
			{Add: WINSINTF_ADD_T{Type: 1, Len: 4, IPAdd: 0x0A000002}, VersNo: dtyp.LARGE_INTEGER(9)},
		},
		MyMaxVersNo:     dtyp.LARGE_INTEGER(9),
		RefreshInterval: 345600,
		WINSStat:        statWithData(),
	})
}

// TestWINSINTF_BROWSER_NAMES_T_RoundTrip exercises the [unique,size_is] array of
// BROWSER_INFO entries, each holding a [unique] ASCII (ndr.STR) name pointer.
func TestWINSINTF_BROWSER_NAMES_T_RoundTrip(t *testing.T) {
	n1, n2 := ndr.STR("ALPHA"), ndr.STR("BETA")
	roundTrip(t, "WINSINTF_BROWSER_NAMES_T", WINSINTF_BROWSER_NAMES_T{
		EntriesRead: 2,
		PInfo: []WINSINTF_BROWSER_INFO_T{
			{DwNameLen: 5, PName: &n1},
			{DwNameLen: 4, PName: &n2},
		},
	})
}

// TestWINSINTF_BIND_DATA_T_RoundTrip exercises the [string] LPSTR members modeled as
// [unique] ASCII string pointers.
func TestWINSINTF_BIND_DATA_T_RoundTrip(t *testing.T) {
	srv, pipe := ndr.STR("192.168.0.1"), ndr.STR(`\pipe\WinsPipe`)
	roundTrip(t, "WINSINTF_BIND_DATA_T", WINSINTF_BIND_DATA_T{
		FTcpIp:     1,
		PServerAdd: &srv,
		PPipeName:  &pipe,
	})
}
