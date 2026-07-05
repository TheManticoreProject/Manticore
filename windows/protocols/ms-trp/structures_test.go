package mstrp

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// ctxHandleHolder embeds a context handle so the round trip drives the same inline
// (no-referent) marshalling path the tapsrv/remotesp request and response stubs use.
type ctxHandleHolder struct {
	Handle PCONTEXT_HANDLE_TYPE
}

// TestPCONTEXT_HANDLE_TYPE_RoundTrip confirms the 20-octet context handle marshals inline
// as exactly 20 bytes and round-trips byte-for-byte.
func TestPCONTEXT_HANDLE_TYPE_RoundTrip(t *testing.T) {
	var in ctxHandleHolder
	for i := range in.Handle {
		in.Handle[i] = byte(i + 1)
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 20 {
		t.Fatalf("wire length = %d, want 20 (a context handle is 20 octets, no referent id)", len(raw))
	}
	if !bytes.Equal(raw, in.Handle[:]) {
		t.Errorf("wire bytes = %x, want %x (handle transmitted inline verbatim)", raw, in.Handle[:])
	}
	var out ctxHandleHolder
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Handle != in.Handle {
		t.Errorf("round-trip mismatch: got %x want %x", out.Handle, in.Handle)
	}
}

// ctxHandle2Holder embeds the remotesp context handle so it, too, is exercised through the
// inline (no-referent) marshalling path.
type ctxHandle2Holder struct {
	Handle PCONTEXT_HANDLE_TYPE2
}

// TestPCONTEXT_HANDLE_TYPE2_RoundTrip confirms the remotesp 20-octet context handle
// marshals inline as exactly 20 bytes and round-trips byte-for-byte.
func TestPCONTEXT_HANDLE_TYPE2_RoundTrip(t *testing.T) {
	var in ctxHandle2Holder
	for i := range in.Handle {
		in.Handle[i] = byte(0xF0 - i)
	}
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(raw) != 20 {
		t.Fatalf("wire length = %d, want 20 (a context handle is 20 octets, no referent id)", len(raw))
	}
	if !bytes.Equal(raw, in.Handle[:]) {
		t.Errorf("wire bytes = %x, want %x (handle transmitted inline verbatim)", raw, in.Handle[:])
	}
	var out ctxHandle2Holder
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Handle != in.Handle {
		t.Errorf("round-trip mismatch: got %x want %x", out.Handle, in.Handle)
	}
}

// TestPCONTEXT_HANDLE_IsZero covers the nulled-out handle the server returns after a
// successful Detach.
func TestPCONTEXT_HANDLE_IsZero(t *testing.T) {
	var zero PCONTEXT_HANDLE_TYPE
	if !zero.IsZero() {
		t.Error("zero-valued PCONTEXT_HANDLE_TYPE should report IsZero()==true")
	}
	var zero2 PCONTEXT_HANDLE_TYPE2
	if !zero2.IsZero() {
		t.Error("zero-valued PCONTEXT_HANDLE_TYPE2 should report IsZero()==true")
	}
	nonzero := PCONTEXT_HANDLE_TYPE{0: 0x01}
	if nonzero.IsZero() {
		t.Error("non-zero PCONTEXT_HANDLE_TYPE should report IsZero()==false")
	}
}
