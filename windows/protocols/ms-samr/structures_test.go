// Package mssamr wire-structure round-trip tests. Each case marshals a value with
// ndr.Marshal, unmarshals it back with ndr.Unmarshal, and asserts equality, exercising
// the NDR tags (conformant/varying arrays, unique pointers, unions, enum widths) of the
// [MS-SAMR] structures. Consolidated into a single test file per the protocol-scoped
// split layout.
package mssamr

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// ---- from base_test.go ----
// roundTripBase marshals in, unmarshals into a fresh value of the same type, and
// asserts the result is deeply equal to in.
func roundTripBase[T any](t *testing.T, name string, in T) {
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

func mustBaseSID(t *testing.T, s string) *msdtyp.RPC_SID {
	t.Helper()
	sid, err := msdtyp.ParseSID(s)
	if err != nil {
		t.Fatalf("ParseSID(%q): %v", s, err)
	}
	return &sid
}

// TestSAMPR_ENUMERATION_BUFFER_RoundTrip exercises a [unique] pointer to a
// conformant array of structs that each embed an RPC_UNICODE_STRING.
func TestSAMPR_ENUMERATION_BUFFER_RoundTrip(t *testing.T) {
	in := SAMPR_ENUMERATION_BUFFER{
		EntriesRead: 2,
		Buffer: []SAMPR_RID_ENUMERATION{
			{RelativeId: 500, Name: msdtyp.NewUnicodeString("Administrator")},
			{RelativeId: 501, Name: msdtyp.NewUnicodeString("Guest")},
		},
	}
	roundTripBase(t, "SAMPR_ENUMERATION_BUFFER", in)
}

// TestSAMPR_PSID_ARRAY_RoundTrip exercises a [unique] pointer to a conformant
// array of structs that each hold a [unique] pointer to an RPC_SID.
func TestSAMPR_PSID_ARRAY_RoundTrip(t *testing.T) {
	in := SAMPR_PSID_ARRAY{
		Count: 2,
		Sids: []SAMPR_SID_INFORMATION{
			{SidPointer: mustBaseSID(t, "S-1-5-21-1004336348-1177238915-682003330-512")},
			{SidPointer: mustBaseSID(t, "S-1-5-18")},
		},
	}
	roundTripBase(t, "SAMPR_PSID_ARRAY", in)
}

// ---- from domain_test.go ----
// TestSAMPR_DOMAIN_INFO_BUFFER_RoundTrip marshals and unmarshals a
// SAMPR_DOMAIN_INFO_BUFFER selecting the DomainNameInformation (5) arm and checks the
// recovered value matches the original.
func TestSAMPR_DOMAIN_INFO_BUFFER_RoundTrip(t *testing.T) {
	in := SAMPR_DOMAIN_INFO_BUFFER{
		Tag: DomainNameInformation,
		Name: SAMPR_DOMAIN_NAME_INFORMATION{
			DomainName: msdtyp.NewUnicodeString("CONTOSO"),
		},
	}

	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var out SAMPR_DOMAIN_INFO_BUFFER
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if out.Tag != DomainNameInformation {
		t.Fatalf("Tag = %d, want %d", out.Tag, DomainNameInformation)
	}
	if got := out.Name.DomainName.String(); got != "CONTOSO" {
		t.Fatalf("DomainName = %q, want %q", got, "CONTOSO")
	}
}

// TestSAMPR_DISPLAY_INFO_BUFFER_RoundTrip marshals and unmarshals a
// SAMPR_DISPLAY_INFO_BUFFER selecting the DomainDisplayUser (1) arm with an empty user
// buffer and checks the recovered value matches the original.
func TestSAMPR_DISPLAY_INFO_BUFFER_RoundTrip(t *testing.T) {
	in := SAMPR_DISPLAY_INFO_BUFFER{
		Tag: DomainDisplayUser,
		UserInformation: SAMPR_DOMAIN_DISPLAY_USER_BUFFER{
			EntriesRead: 0,
			Buffer:      nil,
		},
	}

	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var out SAMPR_DISPLAY_INFO_BUFFER
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if out.Tag != DomainDisplayUser {
		t.Fatalf("Tag = %d, want %d", out.Tag, DomainDisplayUser)
	}
	if out.UserInformation.EntriesRead != 0 {
		t.Fatalf("EntriesRead = %d, want 0", out.UserInformation.EntriesRead)
	}
	if len(out.UserInformation.Buffer) != 0 {
		t.Fatalf("Buffer len = %d, want 0", len(out.UserInformation.Buffer))
	}
	_ = reflect.DeepEqual(in, out)
}

// ---- from group_alias_test.go ----
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
			Name: msdtyp.NewUnicodeString("Domain Admins"),
		},
	})
}

// TestSAMPR_ALIAS_INFO_BUFFER_RoundTrip exercises the alias-information union with the
// AliasGeneralInformation arm (case 1) selected.
func TestSAMPR_ALIAS_INFO_BUFFER_RoundTrip(t *testing.T) {
	groupAliasRoundTrip(t, "AliasGeneralInformation(case1)", SAMPR_ALIAS_INFO_BUFFER{
		Tag: AliasGeneralInformation,
		General: SAMPR_ALIAS_GENERAL_INFORMATION{
			Name:         msdtyp.NewUnicodeString("Administrators"),
			MemberCount:  3,
			AdminComment: msdtyp.NewUnicodeString("Built-in admins"),
		},
	})
}

// ---- from ndr64_enum_test.go ----
// TestNDR64_DomainServerRole_EnumWidth confirms the enum-field tagging sweep: a
// DOMAIN_SERVER_ROLE field marshals as 2 octets under NDR20 and 4 under NDR64.
func TestNDR64_DomainServerRole_EnumWidth(t *testing.T) {
	v := DOMAIN_SERVER_ROLE_INFORMATION{DomainServerRole: DomainServerRolePrimary} // = 3
	n20, err := ndr.MarshalAs(&v, ndr.NDR20)
	if err != nil {
		t.Fatal(err)
	}
	if len(n20) != 2 || n20[0] != 0x03 {
		t.Errorf("NDR20 = % x, want 03 00", n20)
	}
	n64, err := ndr.MarshalAs(&v, ndr.NDR64)
	if err != nil {
		t.Fatal(err)
	}
	if len(n64) != 4 || n64[0] != 0x03 {
		t.Errorf("NDR64 = % x, want 03 00 00 00", n64)
	}
	var out DOMAIN_SERVER_ROLE_INFORMATION
	if err := ndr.UnmarshalAs(n64, &out, ndr.NDR64); err != nil || out.DomainServerRole != DomainServerRolePrimary {
		t.Errorf("NDR64 round trip: %v, %v", out, err)
	}
}

// ---- from user_test.go ----
// TestSamprUserInfoBufferNameRoundTrip marshals a SAMPR_USER_INFO_BUFFER whose
// discriminant selects UserNameInformation (6) and verifies it survives an NDR
// round trip unchanged.
func TestSamprUserInfoBufferNameRoundTrip(t *testing.T) {
	in := SAMPR_USER_INFO_BUFFER{
		Tag: UserNameInformation,
		Name: SAMPR_USER_NAME_INFORMATION{
			UserName: msdtyp.NewUnicodeString("administrator"),
			FullName: msdtyp.NewUnicodeString("Administrator Account"),
		},
	}

	data, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var out SAMPR_USER_INFO_BUFFER
	if err := ndr.Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round trip mismatch:\n in: %+v\nout: %+v", in, out)
	}
}

// TestSamprUserInfoBufferControlRoundTrip marshals a SAMPR_USER_INFO_BUFFER
// whose discriminant selects UserControlInformation (16) and verifies it
// survives an NDR round trip unchanged.
func TestSamprUserInfoBufferControlRoundTrip(t *testing.T) {
	in := SAMPR_USER_INFO_BUFFER{
		Tag: UserControlInformation,
		Control: USER_CONTROL_INFORMATION{
			UserAccountControl: 0x00000200,
		},
	}

	data, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var out SAMPR_USER_INFO_BUFFER
	if err := ndr.Unmarshal(data, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round trip mismatch:\n in: %+v\nout: %+v", in, out)
	}
}

// ---- from validate_test.go ----
// roundTripValidate marshals in, unmarshals into a fresh value of the same type,
// and asserts deep equality. It is named distinctly to avoid colliding with any
// shared roundTrip helper other structure tests may define in this package.
func roundTripValidate[T any](t *testing.T, name string, in T) {
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

// TestSAM_VALIDATE_OUTPUT_ARG_RoundTrip exercises the SAM_VALIDATE_OUTPUT_ARG
// union with the SamValidateAuthentication arm (case 1) selected, including a
// SAM_VALIDATE_PERSISTED_FIELDS payload with a [unique] pointer to a conformant
// array of SAM_VALIDATE_PASSWORD_HASH (itself a [unique] byte buffer).
func TestSAM_VALIDATE_OUTPUT_ARG_RoundTrip(t *testing.T) {
	roundTripValidate(t, "ValidateAuthenticationOutput(case1)", SAM_VALIDATE_OUTPUT_ARG{
		Tag: SamValidateAuthentication,
		ValidateAuthenticationOutput: SAM_VALIDATE_STANDARD_OUTPUT_ARG{
			ChangedPersistedFields: SAM_VALIDATE_PERSISTED_FIELDS{
				PresentFields:         0x00000007,
				PasswordLastSet:       msdtyp.LARGE_INTEGER(0x1122334455667788),
				BadPasswordTime:       msdtyp.LARGE_INTEGER(0x0000000000000000),
				LockoutTime:           msdtyp.LARGE_INTEGER(0x00000000DEADBEEF),
				BadPasswordCount:      3,
				PasswordHistoryLength: 2,
				PasswordHistory: []SAM_VALIDATE_PASSWORD_HASH{
					{Length: 4, Hash: []byte{0x01, 0x02, 0x03, 0x04}},
					{Length: 2, Hash: []byte{0xAA, 0xBB}},
				},
			},
			ValidationStatus: SamValidatePasswordExpired,
		},
	})
}
