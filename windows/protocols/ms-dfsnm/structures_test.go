package msdfsnm

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts
// the result is deeply equal to in. This is the wire-shape acceptance gate for the
// MS-DFSNM (netdfs) NDR structures in the absence of a live DFS server.
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

// TestDFS_TARGET_PRIORITY_CLASS_Width confirms the [v1_enum] priority class marshals
// as a 4-octet value (MS-RPCE 2.2.5.1) — not the 2-octet default NDR enum — and that
// its negative member DfsInvalidPriorityClass (-1) survives the round trip. A 16-bit
// model would truncate the value and desynchronize every following field on the wire.
func TestDFS_TARGET_PRIORITY_CLASS_Width(t *testing.T) {
	type holder struct{ Class DFS_TARGET_PRIORITY_CLASS }
	for _, v := range []DFS_TARGET_PRIORITY_CLASS{
		DfsInvalidPriorityClass,
		DfsSiteCostNormalPriorityClass,
		DfsGlobalHighPriorityClass,
		DfsSiteCostLowPriorityClass,
	} {
		raw, err := ndr.Marshal(&holder{Class: v})
		if err != nil {
			t.Fatalf("Marshal(%d): %v", v, err)
		}
		if len(raw) != 4 {
			t.Fatalf("DFS_TARGET_PRIORITY_CLASS(%d) marshalled to %d bytes, want 4 (v1_enum)", v, len(raw))
		}
		var out holder
		if err := ndr.Unmarshal(raw, &out); err != nil {
			t.Fatalf("Unmarshal(%d): %v", v, err)
		}
		if out.Class != v {
			t.Fatalf("DFS_TARGET_PRIORITY_CLASS round trip: got %d want %d", out.Class, v)
		}
	}
}

// TestDFS_TARGET_PRIORITY exercises the fixed-layout struct that embeds the v1_enum
// priority class followed by two 16-bit fields.
func TestDFS_TARGET_PRIORITY(t *testing.T) {
	roundTrip(t, "DFS_TARGET_PRIORITY", DFS_TARGET_PRIORITY{
		TargetPriorityClass: DfsGlobalHighPriorityClass,
		TargetPriorityRank:  7,
		Reserved:            0,
	})
}

// TestDFS_STORAGE_INFO_1 exercises a struct carrying two [unique] string pointers plus
// the embedded DFS_TARGET_PRIORITY (with its v1_enum), both populated and NULL.
func TestDFS_STORAGE_INFO_1(t *testing.T) {
	roundTrip(t, "DFS_STORAGE_INFO_1/full", DFS_STORAGE_INFO_1{
		State:          2,
		ServerName:     wstr("SERVER01"),
		ShareName:      wstr("public"),
		TargetPriority: DFS_TARGET_PRIORITY{TargetPriorityClass: DfsSiteCostNormalPriorityClass, TargetPriorityRank: 1},
	})
	roundTrip(t, "DFS_STORAGE_INFO_1/nil", DFS_STORAGE_INFO_1{State: 0, ServerName: nil, ShareName: nil})
}

// TestDFS_INFO_3 exercises the [unique] pointer to a conformant array (Storage, sized by
// NumberOfStorages) of DFS_STORAGE_INFO structs, each carrying [unique] string pointers.
func TestDFS_INFO_3(t *testing.T) {
	roundTrip(t, "DFS_INFO_3/two", DFS_INFO_3{
		EntryPath:        wstr(`\\contoso\dfs\folder`),
		Comment:          wstr("a folder"),
		State:            1,
		NumberOfStorages: 2,
		Storage: []DFS_STORAGE_INFO{
			{State: 1, ServerName: wstr("SRV1"), ShareName: wstr("s1")},
			{State: 2, ServerName: wstr("SRV2"), ShareName: wstr("s2")},
		},
	})
	// A NULL storage pointer (no targets yet) decodes back to a nil slice.
	roundTrip(t, "DFS_INFO_3/none", DFS_INFO_3{EntryPath: wstr("x"), NumberOfStorages: 0, Storage: nil})
}

// TestDFS_INFO_9 exercises the richest storage record: a [unique] counted byte blob
// (the self-relative security descriptor, sized by SecurityDescriptorLength) alongside a
// [unique] conformant array of DFS_STORAGE_INFO_1 (each with a v1_enum priority class).
func TestDFS_INFO_9(t *testing.T) {
	roundTrip(t, "DFS_INFO_9", DFS_INFO_9{
		EntryPath:                wstr(`\\contoso\dfs`),
		Comment:                  wstr("root"),
		State:                    1,
		Timeout:                  300,
		Guid:                     guid.GUID{A: 0x11223344, B: 0x5566, C: 0x7788, D: 0x99aa, E: 0xbbccddeeff00},
		PropertyFlags:            0,
		MetadataSize:             4096,
		SecurityDescriptorLength: 4,
		PSecurityDescriptor:      []uint8{0x01, 0x00, 0x04, 0x80},
		NumberOfStorages:         1,
		Storage: []DFS_STORAGE_INFO_1{
			{State: 2, ServerName: wstr("SRV1"), ShareName: wstr("s1"), TargetPriority: DFS_TARGET_PRIORITY{TargetPriorityClass: DfsGlobalHighPriorityClass}},
		},
	})
}

// TestDFS_INFO_50 and the version-info record exercise the unsigned __int64 capability
// fields (8 octets each), the only 64-bit scalars in the protocol.
func TestDFS_INFO_50(t *testing.T) {
	roundTrip(t, "DFS_INFO_50", DFS_INFO_50{
		NamespaceMajorVersion: 2,
		NamespaceMinorVersion: 0,
		NamespaceCapabilities: 0x00000000DEADBEEF,
	})
	roundTrip(t, "DFS_SUPPORTED_NAMESPACE_VERSION_INFO", DFS_SUPPORTED_NAMESPACE_VERSION_INFO{
		DomainDfsMajorVersion:     2,
		DomainDfsMinorVersion:     0,
		DomainDfsCapabilities:     0x0000000100000002,
		StandaloneDfsMajorVersion: 4,
		StandaloneDfsMinorVersion: 1,
		StandaloneDfsCapabilities: 0x0000000300000004,
	})
}

// TestDFSM_ROOT_LIST exercises the INLINE conformant array member (Entry[], embedded —
// not a pointer): its maximum_count is hoisted to the head of the structure and there is
// no referent id. An empty list is cEntries 0 with a zero-length slice.
func TestDFSM_ROOT_LIST(t *testing.T) {
	roundTrip(t, "DFSM_ROOT_LIST/two", DFSM_ROOT_LIST{
		CEntries: 2,
		Entry: []DFSM_ROOT_LIST_ENTRY{
			{ServerShare: wstr(`\\SRV1\share`)},
			{ServerShare: wstr(`\\SRV2\share`)},
		},
	})
	roundTrip(t, "DFSM_ROOT_LIST/empty", DFSM_ROOT_LIST{CEntries: 0, Entry: []DFSM_ROOT_LIST_ENTRY{}})
}

// TestDFS_INFO_STRUCT exercises the top-level [switch_type(unsigned long)] union that
// NetrDfsGetInfo returns: the 4-byte discriminant is transmitted inline ahead of the
// selected [unique] arm. Two arms of different levels are checked.
func TestDFS_INFO_STRUCT(t *testing.T) {
	roundTrip(t, "DFS_INFO_STRUCT/case3", DFS_INFO_STRUCT{
		Tag: 3,
		DfsInfo3: &DFS_INFO_3{
			EntryPath:        wstr(`\\contoso\dfs\f`),
			State:            1,
			NumberOfStorages: 1,
			Storage:          []DFS_STORAGE_INFO{{State: 1, ServerName: wstr("SRV1"), ShareName: wstr("s")}},
		},
	})
	roundTrip(t, "DFS_INFO_STRUCT/case100", DFS_INFO_STRUCT{
		Tag:        100,
		DfsInfo100: &DFS_INFO_100{Comment: wstr("hello")},
	})
}

// TestDFS_INFO_ENUM_STRUCT exercises the nested container union NetrDfsEnum returns:
// an outer Level plus a [switch_is(Level)] union selecting a *_CONTAINER arm whose
// Buffer is a [unique] conformant array of level records.
func TestDFS_INFO_ENUM_STRUCT(t *testing.T) {
	roundTrip(t, "DFS_INFO_ENUM_STRUCT/case3", DFS_INFO_ENUM_STRUCT{
		Level: 3,
		DfsInfoContainer: DFS_INFO_ENUM_STRUCT_DfsInfoContainer{
			Tag: 3,
			DfsInfo3Container: &DFS_INFO_3_CONTAINER{
				EntriesRead: 1,
				Buffer: []DFS_INFO_3{
					{EntryPath: wstr(`\\contoso\dfs\f`), State: 1, NumberOfStorages: 0, Storage: nil},
				},
			},
		},
	})
	// Request shape: Level set, container pointer NULL (no data yet).
	roundTrip(t, "DFS_INFO_ENUM_STRUCT/nil", DFS_INFO_ENUM_STRUCT{
		Level:            1,
		DfsInfoContainer: DFS_INFO_ENUM_STRUCT_DfsInfoContainer{Tag: 1, DfsInfo1Container: nil},
	})
}
