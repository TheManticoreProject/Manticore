package msrpcl

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
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

func wstr(s string) *ndr.WSTR {
	w := ndr.WSTR(s)
	return &w
}

// TestRPCVersion_RoundTrip exercises RPC_VERSION ([MS-RPCL] 2.2): two 16-bit words.
func TestRPCVersion_RoundTrip(t *testing.T) {
	in := RPC_VERSION{MajorVersion: 1, MinorVersion: 0}
	var out RPC_VERSION
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("RPC_VERSION round-trip: got %+v want %+v", out, in)
	}
}

// TestRPCSyntaxIdentifier_RoundTrip exercises RPC_SYNTAX_IDENTIFIER ([MS-RPCL] 2.2): a GUID
// followed by an embedded RPC_VERSION.
func TestRPCSyntaxIdentifier_RoundTrip(t *testing.T) {
	in := RPC_SYNTAX_IDENTIFIER{
		SyntaxGUID:    guid.GUID{A: 0x8a885d04, B: 0x1ceb, C: 0x11c9, D: 0x9fe8, E: 0x08002b104860},
		SyntaxVersion: RPC_VERSION{MajorVersion: 2, MinorVersion: 0},
	}
	var out RPC_SYNTAX_IDENTIFIER
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("RPC_SYNTAX_IDENTIFIER round-trip: got %+v want %+v", out, in)
	}
}

// TestNSIBinding_RoundTrip exercises NSI_BINDING_T ([MS-RPCL] 2.2): two [unique] wide-string
// pointers around a DWORD. Wrapped so the walker sees a top-level struct, as in a response.
func TestNSIBinding_RoundTrip(t *testing.T) {
	type wrap struct{ B NSI_BINDING_T }
	in := wrap{B: NSI_BINDING_T{
		String:            wstr(`ncacn_ip_tcp:10.0.0.1[49152]`),
		Entry_name_syntax: 3,
		Entry_name:        wstr(`/.:/SomeEntry`),
	}}
	var out wrap
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("NSI_BINDING_T round-trip: got %+v want %+v", out, in)
	}
}

// TestNSIBinding_NullStrings covers the null [unique] string pointers (both fields NULL).
func TestNSIBinding_NullStrings(t *testing.T) {
	type wrap struct{ B NSI_BINDING_T }
	in := wrap{B: NSI_BINDING_T{Entry_name_syntax: 3}}
	var out wrap
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("NSI_BINDING_T (null strings) round-trip: got %+v want %+v", out, in)
	}
}

// TestNSIBindingVector_RoundTrip exercises NSI_BINDING_VECTOR_T ([MS-RPCL] 2.2) behind a
// [unique] pointer (NSI_BINDING_VECTOR_P_T), exactly as it appears in the I_nsi_lookup_next
// response: the count word plus the embedded conformant array of NSI_BINDING_T values.
func TestNSIBindingVector_RoundTrip(t *testing.T) {
	type wrap struct {
		V      NSI_BINDING_VECTOR_P_T `ndr:"unique"`
		Status uint16
	}
	in := wrap{
		V: &NSI_BINDING_VECTOR_T{
			Count: 2,
			Binding: []NSI_BINDING_T{
				{String: wstr(`ncacn_np:\\host[\pipe\x]`), Entry_name_syntax: 3, Entry_name: wstr(`/.:/A`)},
				{String: wstr(`ncacn_ip_tcp:10.0.0.2[49153]`), Entry_name_syntax: 3, Entry_name: nil},
			},
		},
		Status: 0,
	}
	var out wrap
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("NSI_BINDING_VECTOR_T round-trip: got %+v want %+v", out, in)
	}
}

// TestNSIBindingVector_Empty covers the empty vector (count 0, zero-length array).
func TestNSIBindingVector_Empty(t *testing.T) {
	type wrap struct {
		V NSI_BINDING_VECTOR_P_T `ndr:"unique"`
	}
	in := wrap{V: &NSI_BINDING_VECTOR_T{Count: 0, Binding: []NSI_BINDING_T{}}}
	var out wrap
	roundTrip(t, &in, &out)
	if out.V == nil || out.V.Count != 0 {
		t.Fatalf("NSI_BINDING_VECTOR_T (empty) round-trip: got %+v", out.V)
	}
}

// TestNSIUUIDVector_RoundTrip exercises NSI_UUID_VECTOR_T ([MS-RPCL] 2.2) behind a [unique]
// pointer (NSI_UUID_VECTOR_P_T), as in the I_nsi_entry_object_inq_next response: the count
// word plus the embedded conformant array of [unique] GUID pointers (NSI_UUID_P_T).
func TestNSIUUIDVector_RoundTrip(t *testing.T) {
	g1 := guid.GUID{A: 0x11111111, B: 0x2222, C: 0x3333, D: 0x4444, E: 0x555566667777}
	g2 := guid.GUID{A: 0xaaaaaaaa, B: 0xbbbb, C: 0xcccc, D: 0xdddd, E: 0xeeeeffff0000}
	type wrap struct {
		V      NSI_UUID_VECTOR_P_T `ndr:"unique"`
		Status uint16
	}
	in := wrap{
		V: &NSI_UUID_VECTOR_T{
			Count: 2,
			Uuid:  []NSI_UUID_P_T{&g1, &g2},
		},
	}
	var out wrap
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("NSI_UUID_VECTOR_T round-trip: got %+v want %+v", out, in)
	}
}

// TestNSINSHandle_RoundTrip exercises NSI_NS_HANDLE_T ([MS-RPCL] 2.2): a 20-byte RPC context
// handle, transmitted inline. Wrapped so the walker sees a top-level struct.
func TestNSINSHandle_RoundTrip(t *testing.T) {
	type wrap struct{ H NSI_NS_HANDLE_T }
	var in wrap
	for i := range in.H {
		in.H[i] = byte(i + 1)
	}
	var out wrap
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("NSI_NS_HANDLE_T round-trip: got %+v want %+v", out, in)
	}
	if len(NSI_NS_HANDLE_T{}) != 20 {
		t.Fatalf("NSI_NS_HANDLE_T must be 20 bytes, got %d", len(NSI_NS_HANDLE_T{}))
	}
}
