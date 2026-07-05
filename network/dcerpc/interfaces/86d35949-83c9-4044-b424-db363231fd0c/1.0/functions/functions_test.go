package functions

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	mstsch "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsch"
)

// roundTrip marshals in, unmarshals into a fresh T, and asserts deep equality. It
// validates the NDR tags on the hand-tuned ITaskSchedulerService [out] response shapes —
// the double-pointer out strings, the [unique] pointers to conformant arrays of GUIDs /
// SYSTEMTIMEs / strings — which are not exercised by a live server in CI.
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

// TestEnumInstancesResponse exercises a [unique] pointer to a conformant array of GUID
// values (GUID** with size_is(,*pcGuids)).
func TestEnumInstancesResponse(t *testing.T) {
	roundTrip(t, "EnumInstances", schRpcEnumInstancesResponse{
		PcGuids: 2,
		PGuids: []guid.GUID{
			{A: 0x11111111, B: 0x2222, C: 0x3333, D: 0x4444, E: 0x555566667777},
			{A: 0x89abcdef, B: 0x0123, C: 0x4567, D: 0x89ab, E: 0xcdef01234567},
		},
		Status: 0,
	})
	roundTrip(t, "EnumInstances/empty", schRpcEnumInstancesResponse{PcGuids: 0, PGuids: []guid.GUID{}})
}

// TestEnumFoldersResponse exercises a [unique] pointer to a conformant array of [unique]
// wide-string pointers (TASK_NAMES* with size_is(,*pcNames)), plus an [in,out] index.
func TestEnumFoldersResponse(t *testing.T) {
	roundTrip(t, "EnumFolders", schRpcEnumFoldersResponse{
		PStartIndex: 3,
		PcNames:     2,
		PNames:      []*ndr.WSTR{wstr("\\Microsoft"), wstr("\\Custom Tasks")},
		Status:      1, // S_FALSE — more entries
	})
}

// TestScheduledRuntimesResponse exercises a [unique] pointer to a conformant array of
// SYSTEMTIME values (PSYSTEMTIME* with size_is(,*pcRuntimes)).
func TestScheduledRuntimesResponse(t *testing.T) {
	roundTrip(t, "ScheduledRuntimes", schRpcScheduledRuntimesResponse{
		PcRuntimes: 2,
		PRuntimes: []dtyp.SYSTEMTIME{
			{WYear: 2026, WMonth: 7, WDay: 5, WHour: 9, WMinute: 30},
			{WYear: 2026, WMonth: 7, WDay: 6, WHour: 12},
		},
	})
}

// TestGetInstanceInfoResponse exercises the mixed shape: three [out] double-pointer
// strings, a [unique] GUID array, and inline scalars.
func TestGetInstanceInfoResponse(t *testing.T) {
	roundTrip(t, "GetInstanceInfo", schRpcGetInstanceInfoResponse{
		PPath:            wstr("\\Task1"),
		PState:           3,
		PCurrentAction:   wstr("exec"),
		PInfo:            nil,
		PcGroupInstances: 1,
		PGroupInstances:  []guid.GUID{{A: 0xdeadbeef, B: 1, C: 2, D: 3, E: 4}},
		PEnginePID:       4242,
	})
}

// TestRegisterTaskRequest exercises the [in] side: a [unique] pointer to a conformant
// array of TASK_USER_CRED structs (each with two [unique] strings) plus [unique] strings.
func TestRegisterTaskRequest(t *testing.T) {
	roundTrip(t, "RegisterTask", schRpcRegisterTaskRequest{
		Path:      wstr("\\NewTask"),
		Xml:       ndr.WSTR("<Task/>"),
		Flags:     0,
		Sddl:      nil,
		LogonType: 1,
		CCreds:    1,
		PCreds: []mstsch.TASK_USER_CRED{
			{UserId: wstr("DOMAIN\\svc"), Password: wstr("pw"), Flags: 1},
		},
	})
}
