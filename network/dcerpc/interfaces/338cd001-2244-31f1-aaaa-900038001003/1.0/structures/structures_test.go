package structures

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// TestRPCHKEYSize pins the 20-byte context-handle representation ([MS-RPCE] 2.3.2.2).
func TestRPCHKEYSize(t *testing.T) {
	var h RPC_HKEY
	if !h.IsZero() {
		t.Fatal("zero-value handle should be zero")
	}
	raw, err := ndr.Marshal(&struct{ H RPC_HKEY }{})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 20 {
		t.Errorf("RPC_HKEY marshalled to %d bytes, want 20", len(raw))
	}
}

// TestRVALENTRoundTrip exercises a struct with a [unique] RPC_UNICODE_STRING pointer and a
// [unique] DWORD pointer (the value_ent record).
func TestRVALENTRoundTrip(t *testing.T) {
	name := dtyp.NewUnicodeString("Version")
	ptr := ndr.DWORD(0x40)
	in := RVALENT{Ve_valuename: &name, Ve_valuelen: 8, Ve_valueptr: &ptr, Ve_type: 1}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out RVALENT
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Ve_valuename == nil || out.Ve_valuename.String() != "Version" {
		t.Errorf("Ve_valuename round-trip: %+v", out.Ve_valuename)
	}
	if out.Ve_valueptr == nil || *out.Ve_valueptr != 0x40 {
		t.Errorf("Ve_valueptr round-trip: %v", out.Ve_valueptr)
	}
	if out.Ve_valuelen != 8 || out.Ve_type != 1 {
		t.Errorf("scalar round-trip: len=%d type=%d", out.Ve_valuelen, out.Ve_type)
	}
}

// TestRPCSecurityDescriptorRoundTrip exercises a [unique] pointer to a conformant-varying
// byte array whose maximum_count/actual_count come from sibling DWORDs (size_is /
// length_is), with maximum_count > actual_count (an over-allocated buffer).
func TestRPCSecurityDescriptorRoundTrip(t *testing.T) {
	in := RPC_SECURITY_DESCRIPTOR{
		LpSecurityDescriptor:    []uint8{1, 2, 3, 4},
		CbInSecurityDescriptor:  8, // capacity
		CbOutSecurityDescriptor: 4, // valid bytes
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out RPC_SECURITY_DESCRIPTOR
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(out.LpSecurityDescriptor, []uint8{1, 2, 3, 4}) {
		t.Errorf("LpSecurityDescriptor round-trip: %v", out.LpSecurityDescriptor)
	}
}

// TestRPCSecurityAttributesRoundTrip exercises a struct embedding RPC_SECURITY_DESCRIPTOR
// (itself a pointer-bearing struct) followed by a 1-octet NDR BOOLEAN.
func TestRPCSecurityAttributesRoundTrip(t *testing.T) {
	in := RPC_SECURITY_ATTRIBUTES{
		NLength:               12,
		RpcSecurityDescriptor: RPC_SECURITY_DESCRIPTOR{LpSecurityDescriptor: []uint8{9}, CbInSecurityDescriptor: 1, CbOutSecurityDescriptor: 1},
		BInheritHandle:        true,
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out RPC_SECURITY_ATTRIBUTES
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.NLength != 12 || !out.BInheritHandle ||
		!reflect.DeepEqual(out.RpcSecurityDescriptor.LpSecurityDescriptor, []uint8{9}) {
		t.Errorf("RPC_SECURITY_ATTRIBUTES round-trip: %+v", out)
	}
}
