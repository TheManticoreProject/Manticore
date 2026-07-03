package mspcq

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals v then unmarshals into out (both pointers to the same type).
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

// TestRPCHQuery_RoundTrip exercises RPC_HQUERY ([MS-PCQ] 2.2.1): a 20-byte RPC context
// handle. It is wrapped in a struct so the walker sees a top-level parameter struct, as it
// would in a request/response.
func TestRPCHQuery_RoundTrip(t *testing.T) {
	type wrap struct{ H RPC_HQUERY }
	var in wrap
	for i := range in.H {
		in.H[i] = byte(i + 1)
	}
	var out wrap
	roundTrip(t, &in, &out)
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("RPC_HQUERY round-trip: got %+v want %+v", out, in)
	}
	if len(RPC_HQUERY{}) != 20 {
		t.Fatalf("RPC_HQUERY must be 20 bytes, got %d", len(RPC_HQUERY{}))
	}
}

// TestRPCHQuery_IsZero verifies the zero-handle predicate used after a successful
// PerflibV2CloseQueryHandle (the server zeroes the handle on return).
func TestRPCHQuery_IsZero(t *testing.T) {
	var zero RPC_HQUERY
	if !zero.IsZero() {
		t.Fatal("zero-valued RPC_HQUERY: IsZero() = false, want true")
	}
	nonZero := RPC_HQUERY{0: 1}
	if nonZero.IsZero() {
		t.Fatal("non-zero RPC_HQUERY: IsZero() = true, want false")
	}
}

// TestPRPCHQuery_Alias confirms PRPC_HQUERY is an alias of RPC_HQUERY (the IDL declares
// PRPC_HQUERY as RPC_HQUERY*, transmitted as the same 20-byte context handle inline).
func TestPRPCHQuery_Alias(t *testing.T) {
	if reflect.TypeOf(PRPC_HQUERY{}) != reflect.TypeOf(RPC_HQUERY{}) {
		t.Fatal("PRPC_HQUERY must be an alias of RPC_HQUERY")
	}
}
