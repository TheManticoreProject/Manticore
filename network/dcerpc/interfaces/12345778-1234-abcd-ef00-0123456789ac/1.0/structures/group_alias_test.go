package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// groupAliasRoundTrip marshals in, unmarshals into a fresh T, and asserts equality.
func groupAliasRoundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()

	raw, err := ndr.Marshal(in)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}

	var out T
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("%s: round-trip mismatch\n in: %#v\nout: %#v", name, in, out)
	}
}

// TestSAMPR_GROUP_INFO_BUFFER_RoundTrip exercises the group-information union with the
// GroupNameInformation arm (case 2) selected.
func TestSAMPR_GROUP_INFO_BUFFER_RoundTrip(t *testing.T) {
	groupAliasRoundTrip(t, "GroupNameInformation(case2)", SAMPR_GROUP_INFO_BUFFER{
		Tag: GroupNameInformation,
		Name: SAMPR_GROUP_NAME_INFORMATION{
			Name: dtyp.NewUnicodeString("Domain Admins"),
		},
	})
}

// TestSAMPR_ALIAS_INFO_BUFFER_RoundTrip exercises the alias-information union with the
// AliasGeneralInformation arm (case 1) selected.
func TestSAMPR_ALIAS_INFO_BUFFER_RoundTrip(t *testing.T) {
	groupAliasRoundTrip(t, "AliasGeneralInformation(case1)", SAMPR_ALIAS_INFO_BUFFER{
		Tag: AliasGeneralInformation,
		General: SAMPR_ALIAS_GENERAL_INFORMATION{
			Name:         dtyp.NewUnicodeString("Administrators"),
			MemberCount:  3,
			AdminComment: dtyp.NewUnicodeString("Built-in admins"),
		},
	})
}
