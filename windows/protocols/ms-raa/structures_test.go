package msraa

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals v (a pointer) then unmarshals into out (a pointer to the same type)
// and fails the test on any codec error. Callers compare *v and *out with reflect.DeepEqual.
func roundTrip(t *testing.T, v any, out any) {
	t.Helper()
	raw, err := ndr.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%T): %v", v, err)
	}
	if err := ndr.Unmarshal(raw, out); err != nil {
		t.Fatalf("Unmarshal(%T): %v", v, err)
	}
}

func mustSID(t *testing.T, s string) dtyp.RPC_SID {
	t.Helper()
	sid, err := dtyp.ParseSID(s)
	if err != nil {
		t.Fatalf("ParseSID(%q): %v", s, err)
	}
	return sid
}

// TestAuthzrHandle_RoundTrip exercises AUTHZR_HANDLE ([MS-RAA] 2.2.1.1): the 20-byte RPC
// context handle, transmitted inline. It is wrapped so the walker sees a top-level struct,
// as in a request/response.
func TestAuthzrHandle_RoundTrip(t *testing.T) {
	type wrap struct{ H AUTHZR_HANDLE }
	var in wrap
	for i := range in.H {
		in.H[i] = byte(i + 1)
	}
	var out wrap
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("AUTHZR_HANDLE round-trip: got %+v want %+v", out, in)
	}
	if len(AUTHZR_HANDLE{}) != AuthzrHandleSize {
		t.Fatalf("AUTHZR_HANDLE must be %d bytes, got %d", AuthzrHandleSize, len(AUTHZR_HANDLE{}))
	}
}

// TestAuthzrHandle_IsNull covers the null-handle predicate (the 16-octet UUID after the
// 4-octet attributes word being all zero).
func TestAuthzrHandle_IsNull(t *testing.T) {
	var zero AUTHZR_HANDLE
	if !zero.IsNull() {
		t.Fatal("zero-valued AUTHZR_HANDLE: IsNull() = false, want true")
	}
	// Non-zero only in the attributes word: still the null handle by UUID.
	attrsOnly := AUTHZR_HANDLE{0: 1, 1: 2, 2: 3, 3: 4}
	if !attrsOnly.IsNull() {
		t.Fatal("attributes-only AUTHZR_HANDLE: IsNull() = false, want true")
	}
	nonNull := AUTHZR_HANDLE{4: 1}
	if nonNull.IsNull() {
		t.Fatal("non-null AUTHZR_HANDLE: IsNull() = true, want false")
	}
}

// TestObjectTypeList_RoundTrip covers OBJECT_TYPE_LIST ([MS-DTYP]): WORD Level, ACCESS_MASK
// Remaining, and a [unique] GUID pointer.
func TestObjectTypeList_RoundTrip(t *testing.T) {
	in := OBJECT_TYPE_LIST{
		Level:      1,
		Remaining:  0x00020019,
		ObjectType: &dtyp.GUID{Data1: 0x11223344, Data2: 0x5566, Data3: 0x7788, Data4: [8]byte{1, 2, 3, 4, 5, 6, 7, 8}},
	}
	var out OBJECT_TYPE_LIST
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("OBJECT_TYPE_LIST round-trip: got %+v want %+v", out, in)
	}
}

// TestAuthzrAccessRequest_RoundTrip covers AUTHZR_ACCESS_REQUEST ([MS-RAA] 2.2.3.1): a
// [unique] SID pointer plus a [unique] pointer to a size_is(ObjectTypeListLength) array.
func TestAuthzrAccessRequest_RoundTrip(t *testing.T) {
	sid := mustSID(t, "S-1-5-21-1004336348-1177238915-682003330-512")
	otl := []OBJECT_TYPE_LIST{
		{Level: 0, Remaining: 0x1, ObjectType: &dtyp.GUID{Data1: 1}},
		{Level: 1, Remaining: 0x2, ObjectType: &dtyp.GUID{Data1: 2}},
	}
	in := AUTHZR_ACCESS_REQUEST{
		DesiredAccess:        0x02000000,
		PrincipalSelfSid:     &sid,
		ObjectTypeListLength: ndr.DWORD(len(otl)),
		ObjectTypeList:       otl,
	}
	var out AUTHZR_ACCESS_REQUEST
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("AUTHZR_ACCESS_REQUEST round-trip: got %+v want %+v", out, in)
	}
}

// TestAuthzrAccessReply_RoundTrip covers AUTHZR_ACCESS_REPLY ([MS-RAA] 2.2.3.2): two
// [unique] conformant arrays both sized by ResultListLength.
func TestAuthzrAccessReply_RoundTrip(t *testing.T) {
	in := AUTHZR_ACCESS_REPLY{
		ResultListLength:  3,
		GrantedAccessMask: []ndr.DWORD{0x1, 0x2, 0x3},
		Error:             []ndr.DWORD{0, 5, 0},
	}
	var out AUTHZR_ACCESS_REPLY
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("AUTHZR_ACCESS_REPLY round-trip: got %+v want %+v", out, in)
	}
}

// TestSRSD_RoundTrip covers SR_SD ([MS-RAA] 2.2.3.3): a serialized security-descriptor
// blob as a [unique] size_is(dwLength) byte array.
func TestSRSD_RoundTrip(t *testing.T) {
	blob := make([]uint8, 24)
	for i := range blob {
		blob[i] = byte(i)
	}
	in := SR_SD{DwLength: ndr.DWORD(len(blob)), PSrSd: blob}
	var out SR_SD
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("SR_SD round-trip: got %+v want %+v", out, in)
	}
}

// TestAuthzrTokenGroups_RoundTrip covers AUTHZR_TOKEN_GROUPS and AUTHZR_SID_AND_ATTRIBUTES
// ([MS-RAA] 2.2.3.9 / 2.2.3.8): a size_is(GroupCount) array of SID+attribute pairs.
func TestAuthzrTokenGroups_RoundTrip(t *testing.T) {
	s0 := mustSID(t, "S-1-5-32-544")
	s1 := mustSID(t, "S-1-1-0")
	in := AUTHZR_TOKEN_GROUPS{
		GroupCount: 2,
		Groups: []AUTHZR_SID_AND_ATTRIBUTES{
			{Sid: &s0, Attributes: 0x7},
			{Sid: &s1, Attributes: 0x4},
		},
	}
	var out AUTHZR_TOKEN_GROUPS
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("AUTHZR_TOKEN_GROUPS round-trip: got %+v want %+v", out, in)
	}
}

// TestAuthzrTokenUser_RoundTrip covers AUTHZR_TOKEN_USER ([MS-RAA] 2.2.3.10): a single
// embedded SID-and-attributes value.
func TestAuthzrTokenUser_RoundTrip(t *testing.T) {
	sid := mustSID(t, "S-1-5-21-1-2-3-1105")
	in := AUTHZR_TOKEN_USER{User: AUTHZR_SID_AND_ATTRIBUTES{Sid: &sid, Attributes: 0}}
	var out AUTHZR_TOKEN_USER
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("AUTHZR_TOKEN_USER round-trip: got %+v want %+v", out, in)
	}
}

// TestAuthzrSecurityAttributeStringValue_RoundTrip covers the [string][size_is(Length)]
// WCHAR* conformant-varying string arm ([MS-RAA] 2.2.3.4).
func TestAuthzrSecurityAttributeStringValue_RoundTrip(t *testing.T) {
	val := []uint16{'H', 'e', 'l', 'l', 'o', 0}
	in := AUTHZR_SECURITY_ATTRIBUTE_STRING_VALUE{Length: ndr.DWORD(len(val)), Value: val}
	var out AUTHZR_SECURITY_ATTRIBUTE_STRING_VALUE
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("AUTHZR_SECURITY_ATTRIBUTE_STRING_VALUE round-trip: got %+v want %+v", out, in)
	}
}

// TestAuthzrSecurityAttributeUnion_RoundTrip exercises each arm of the
// switch_is(ValueType) union in AUTHZR_SECURITY_ATTRIBUTE_V1_VALUE ([MS-RAA] 2.2.3.5),
// including the 0x2/0x6 labels that share the UINT64 arm.
func TestAuthzrSecurityAttributeUnion_RoundTrip(t *testing.T) {
	str := []uint16{'x', 'y', 'z', 0}
	cases := []struct {
		name string
		in   AUTHZR_SECURITY_ATTRIBUTE_V1_VALUE
	}{
		{"int64", AUTHZR_SECURITY_ATTRIBUTE_V1_VALUE{ValueType: 1, AttributeUnion: AUTHZR_SECURITY_ATTRIBUTE_UNION{Tag: 1, Int64: -42}}},
		{"uint64", AUTHZR_SECURITY_ATTRIBUTE_V1_VALUE{ValueType: 2, AttributeUnion: AUTHZR_SECURITY_ATTRIBUTE_UNION{Tag: 2, Uint64: 0xdeadbeefcafe}}},
		{"boolean", AUTHZR_SECURITY_ATTRIBUTE_V1_VALUE{ValueType: 6, AttributeUnion: AUTHZR_SECURITY_ATTRIBUTE_UNION{Tag: 6, Uint64Bool: 1}}},
		{"string", AUTHZR_SECURITY_ATTRIBUTE_V1_VALUE{ValueType: 3, AttributeUnion: AUTHZR_SECURITY_ATTRIBUTE_UNION{Tag: 3, String: AUTHZR_SECURITY_ATTRIBUTE_STRING_VALUE{Length: ndr.DWORD(len(str)), Value: str}}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out AUTHZR_SECURITY_ATTRIBUTE_V1_VALUE
			roundTrip(t, &c.in, &out)
			if !reflect.DeepEqual(c.in, out) {
				t.Fatalf("union %s round-trip: got %+v want %+v", c.name, out, c.in)
			}
		})
	}
}

// TestAuthzrSecurityAttributesInformation_RoundTrip covers the nested
// AUTHZR_SECURITY_ATTRIBUTES_INFORMATION / AUTHZR_SECURITY_ATTRIBUTE_V1 chain ([MS-RAA]
// 2.2.3.6 / 2.2.3.7) with a value array carrying two arms.
func TestAuthzrSecurityAttributesInformation_RoundTrip(t *testing.T) {
	name := []uint16{'A', 'g', 'e', 0}
	in := AUTHZR_SECURITY_ATTRIBUTES_INFORMATION{
		Version:        1,
		Reserved:       0,
		AttributeCount: 1,
		Attributes: []AUTHZR_SECURITY_ATTRIBUTE_V1{
			{
				Length:     ndr.DWORD(len(name)),
				Value:      name,
				ValueType:  1,
				Reserved:   0,
				Flags:      0,
				ValueCount: 2,
				Values: []AUTHZR_SECURITY_ATTRIBUTE_V1_VALUE{
					{ValueType: 1, AttributeUnion: AUTHZR_SECURITY_ATTRIBUTE_UNION{Tag: 1, Int64: 40}},
					{ValueType: 1, AttributeUnion: AUTHZR_SECURITY_ATTRIBUTE_UNION{Tag: 1, Int64: 41}},
				},
			},
		},
	}
	var out AUTHZR_SECURITY_ATTRIBUTES_INFORMATION
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("AUTHZR_SECURITY_ATTRIBUTES_INFORMATION round-trip: got %+v want %+v", out, in)
	}
}

// TestAuthzrContextInformation_RoundTrip exercises the switch_is(ValueType) union of
// AUTHZR_CONTEXT_INFORMATION ([MS-RAA] 2.2.3.7) for the user, groups, and claims arms.
func TestAuthzrContextInformation_RoundTrip(t *testing.T) {
	uSid := mustSID(t, "S-1-5-21-9-8-7-500")
	gSid := mustSID(t, "S-1-5-32-544")
	cases := []struct {
		name string
		in   AUTHZR_CONTEXT_INFORMATION
	}{
		{"user", AUTHZR_CONTEXT_INFORMATION{
			ValueType: 1,
			ContextInfoUnion: AUTHZR_CONTEXT_INFORMATION_UNION{
				Tag:        1,
				PTokenUser: &AUTHZR_TOKEN_USER{User: AUTHZR_SID_AND_ATTRIBUTES{Sid: &uSid}},
			},
		}},
		{"groups", AUTHZR_CONTEXT_INFORMATION{
			ValueType: 2,
			ContextInfoUnion: AUTHZR_CONTEXT_INFORMATION_UNION{
				Tag:          2,
				PTokenGroups: &AUTHZR_TOKEN_GROUPS{GroupCount: 1, Groups: []AUTHZR_SID_AND_ATTRIBUTES{{Sid: &gSid, Attributes: 0x7}}},
			},
		}},
		{"restricted-sids", AUTHZR_CONTEXT_INFORMATION{
			ValueType: 3,
			ContextInfoUnion: AUTHZR_CONTEXT_INFORMATION_UNION{
				Tag:                 3,
				PTokenRestrictedSid: &AUTHZR_TOKEN_GROUPS{GroupCount: 1, Groups: []AUTHZR_SID_AND_ATTRIBUTES{{Sid: &gSid}}},
			},
		}},
		{"user-claims", AUTHZR_CONTEXT_INFORMATION{
			ValueType: 13,
			ContextInfoUnion: AUTHZR_CONTEXT_INFORMATION_UNION{
				Tag:              13,
				PTokenUserClaims: &AUTHZR_SECURITY_ATTRIBUTES_INFORMATION{Version: 1, AttributeCount: 0},
			},
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out AUTHZR_CONTEXT_INFORMATION
			roundTrip(t, &c.in, &out)
			if !reflect.DeepEqual(c.in, out) {
				t.Fatalf("AUTHZR_CONTEXT_INFORMATION %s round-trip: got %+v want %+v", c.name, out, c.in)
			}
		})
	}
}

// TestEnumValues pins the discriminant enum values the union case tags rely on.
func TestEnumValues(t *testing.T) {
	if AuthzContextInfoUserSid != 1 || AuthzContextInfoGroupsSids != 2 || AuthzContextInfoDeviceClaims != 14 {
		t.Fatalf("AUTHZ_CONTEXT_INFORMATION_CLASS values drifted")
	}
	if AUTHZ_SECURITY_ATTRIBUTE_OPERATION_REPLACE != 4 || AUTHZ_SID_OPERATION_DELETE != 3 {
		t.Fatalf("operation enum values drifted")
	}
}
