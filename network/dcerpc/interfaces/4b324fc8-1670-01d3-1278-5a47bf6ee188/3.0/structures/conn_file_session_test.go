package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTripConnFileSession marshals in, unmarshals into a fresh value of the same
// type, and asserts the result is deeply equal to in. It is named distinctly to
// avoid colliding with helpers defined by other test files in this package.
func roundTripConnFileSession[T any](t *testing.T, name string, in T) {
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

// TestSESSION_INFO_1_CONTAINER_RoundTrip exercises a [unique] pointer to a
// conformant array of structs that each carry [unique] string pointers.
func TestSESSION_INFO_1_CONTAINER_RoundTrip(t *testing.T) {
	in := SESSION_INFO_1_CONTAINER{
		EntriesRead: 2,
		Buffer: []SESSION_INFO_1{
			{
				Sesi1Cname:     "\\\\CLIENT1",
				Sesi1Username:  "alice",
				Sesi1NumOpens:  3,
				Sesi1Time:      120,
				Sesi1IdleTime:  10,
				Sesi1UserFlags: 1,
			},
			{
				Sesi1Cname:     "\\\\CLIENT2",
				Sesi1Username:  "",
				Sesi1NumOpens:  0,
				Sesi1Time:      0,
				Sesi1IdleTime:  0,
				Sesi1UserFlags: 0,
			},
		},
	}
	roundTripConnFileSession(t, "SESSION_INFO_1_CONTAINER", in)
}

// TestSESSION_ENUM_STRUCT_Level1_RoundTrip exercises the enum struct with a level-1
// union arm pointing at a one-element container.
func TestSESSION_ENUM_STRUCT_Level1_RoundTrip(t *testing.T) {
	in := SESSION_ENUM_STRUCT{
		Level: 1,
		SessionInfo: SESSION_ENUM_UNION{
			Tag: 1,
			Level1: &SESSION_INFO_1_CONTAINER{
				EntriesRead: 1,
				Buffer: []SESSION_INFO_1{
					{
						Sesi1Cname:     "\\\\HOST",
						Sesi1Username:  "bob",
						Sesi1NumOpens:  5,
						Sesi1Time:      300,
						Sesi1IdleTime:  42,
						Sesi1UserFlags: 0,
					},
				},
			},
		},
	}
	roundTripConnFileSession(t, "SESSION_ENUM_STRUCT_Level1", in)
}
