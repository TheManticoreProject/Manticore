package msfsrvp

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts
// the result is deeply equal to in. This is the wire-shape acceptance gate for the
// MS-FSRVP (FssagentRpc) NDR structures in the absence of a live File Server VSS agent.
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

func mustGUID(t *testing.T, s string) dtyp.GUID {
	t.Helper()
	g, err := guid.FromString(s)
	if err != nil {
		t.Fatalf("FromString(%q): %v", s, err)
	}
	return dtyp.NewGUID(*g)
}

// TestGUIDFieldWidth pins the wire width: an FSSAGENT_SHARE_MAPPING_1 with NULL strings
// marshals its two dtyp.GUIDs as 16 octets each (32 total) + a 4-byte referent-id word
// for each NULL LPWSTR (null referent id = 0) + 8-byte LONGLONG. This guards against a
// regression back to windows/guid.GUID (which would over-align to 24 octets/GUID).
func TestGUIDFieldWidth(t *testing.T) {
	raw, err := ndr.Marshal(&FSSAGENT_SHARE_MAPPING_1{})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	// 16 + 16 (GUIDs) + 4 + 4 (two null unique referent ids) + 8 (LONGLONG) = 48.
	if len(raw) != 48 {
		t.Fatalf("FSSAGENT_SHARE_MAPPING_1 marshalled to %d bytes, want 48 (16-octet GUIDs); a 24-octet GUID would push this to 64", len(raw))
	}
}

// TestFSSAGENT_SHARE_MAPPING_1 exercises the level-1 share mapping ([MS-FSRVP] 2.2.3.2):
// two GUIDs, two [unique][string] LPWSTR pointers, and a LONGLONG timestamp. Both the
// populated case and the NULL-string case (the [unique] referents absent) must survive.
func TestFSSAGENT_SHARE_MAPPING_1(t *testing.T) {
	roundTrip(t, "FSSAGENT_SHARE_MAPPING_1/populated", FSSAGENT_SHARE_MAPPING_1{
		ShadowCopySetId:     mustGUID(t, "11111111-2222-3333-4444-555555555555"),
		ShadowCopyId:        mustGUID(t, "66666666-7777-8888-9999-aaaaaaaaaaaa"),
		ShareNameUNC:        wstr(`\\server\share`),
		ShadowCopyShareName: wstr(`\\server\share@{GMT}`),
		CreationTimestamp:   133000000000000000,
	})

	roundTrip(t, "FSSAGENT_SHARE_MAPPING_1/nil-strings", FSSAGENT_SHARE_MAPPING_1{
		ShadowCopySetId:     mustGUID(t, "11111111-2222-3333-4444-555555555555"),
		ShadowCopyId:        mustGUID(t, "66666666-7777-8888-9999-aaaaaaaaaaaa"),
		ShareNameUNC:        nil,
		ShadowCopyShareName: nil,
		CreationTimestamp:   0,
	})
}

// TestFSSAGENT_SHARE_MAPPING exercises the [switch_type(unsigned long)] union GetShareMapping
// returns ([MS-FSRVP] 2.2.3.1): the 4-byte discriminant inline, then the case(1) arm — a
// [unique] pointer to a level-1 mapping. Covers the selected arm, the NULL-referent request
// shape, and the [default] (empty) arm.
func TestFSSAGENT_SHARE_MAPPING(t *testing.T) {
	roundTrip(t, "FSSAGENT_SHARE_MAPPING/case1", FSSAGENT_SHARE_MAPPING{
		Tag: 1,
		ShareMapping1: &FSSAGENT_SHARE_MAPPING_1{
			ShadowCopySetId:     mustGUID(t, "11111111-2222-3333-4444-555555555555"),
			ShadowCopyId:        mustGUID(t, "66666666-7777-8888-9999-aaaaaaaaaaaa"),
			ShareNameUNC:        wstr(`\\server\share`),
			ShadowCopyShareName: wstr(`\\server\share@{GMT}`),
			CreationTimestamp:   133000000000000000,
		},
	})

	// Level 1 requested, arm pointer NULL (the [unique] referent is absent on the wire).
	roundTrip(t, "FSSAGENT_SHARE_MAPPING/case1-nil", FSSAGENT_SHARE_MAPPING{
		Tag:           1,
		ShareMapping1: nil,
	})

	// [default] arm: any discriminant other than 1 selects the empty arm.
	roundTrip(t, "FSSAGENT_SHARE_MAPPING/default", FSSAGENT_SHARE_MAPPING{
		Tag:           0,
		ShareMapping1: nil,
	})
}
