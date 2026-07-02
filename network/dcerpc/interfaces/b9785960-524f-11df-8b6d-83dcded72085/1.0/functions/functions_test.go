package functions

import (
	"bytes"
	"testing"

	ISDKey "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b9785960-524f-11df-8b6d-83dcded72085/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// captureInvoker records the marshalled request stub and opnum without any network I/O,
// and unmarshals a canned response stub into the out value so both directions of the NDR
// layout can be asserted.
type captureInvoker struct {
	stub  []byte
	opnum uint16
	resp  []byte // canned response stub to unmarshal into out (nil = leave zero)
}

func (c *captureInvoker) Invoke(in ndr.Call, out any) error {
	b, err := ndr.Request(in)
	if err != nil {
		return err
	}
	c.stub = b
	c.opnum = in.Opnum()
	if c.resp != nil {
		return ndr.Response(c.resp, out)
	}
	return nil
}

// TestGetKeyRequestMarshalNoRootKey covers the "latest group key" request shape: a NULL
// pRootKeyID (referent id 0) and L0/L1/L2 all -1. pbTargetSD is a top-level [ref] pointer
// to a conformant char array, so its max_count is emitted inline (no referent id) right
// after cbTargetSD, and no referent data trails the fixed part.
func TestGetKeyRequestMarshalNoRootKey(t *testing.T) {
	cap := &captureInvoker{}
	_, _, _ = GetKey(cap, 4, []byte{0x01, 0x02, 0x03, 0x04}, nil, -1, -1, -1)

	if cap.opnum != ISDKey.OpnumGetKey {
		t.Fatalf("opnum = %d, want %d", cap.opnum, ISDKey.OpnumGetKey)
	}
	want := []byte{
		0x04, 0x00, 0x00, 0x00, // cbTargetSD = 4
		0x04, 0x00, 0x00, 0x00, // pbTargetSD conformant max_count = 4 (ref: no referent id)
		0x01, 0x02, 0x03, 0x04, // the 4 SD octets (already 4-aligned)
		0x00, 0x00, 0x00, 0x00, // pRootKeyID unique referent id = 0 (NULL)
		0xff, 0xff, 0xff, 0xff, // L0KeyID = -1
		0xff, 0xff, 0xff, 0xff, // L1KeyID = -1
		0xff, 0xff, 0xff, 0xff, // L2KeyID = -1
	}
	if !bytes.Equal(cap.stub, want) {
		t.Fatalf("request stub = % x, want % x", cap.stub, want)
	}
}

// TestGetKeyRequestMarshalWithRootKey checks that a non-NULL pRootKeyID is carried as the
// 16-octet [MS-DTYP] wire GUID (Data1/2/3 little-endian, Data4 verbatim) behind a non-zero
// referent id — not windows/guid.GUID's 24-octet reflected layout. The layout is
// cbTargetSD(4) | pbTargetSD max_count(4) | referent id(4) | GUID(16) | L0(4) | L1(4) | L2(4).
func TestGetKeyRequestMarshalWithRootKey(t *testing.T) {
	g, err := guid.FromString("b9785960-524f-11df-8b6d-83dcded72085")
	if err != nil {
		t.Fatalf("parse guid: %v", err)
	}
	cap := &captureInvoker{}
	_, _, _ = GetKey(cap, 0, []byte{}, g, 0, 1, 2)

	want := []byte{
		0x00, 0x00, 0x00, 0x00, // cbTargetSD = 0
		0x00, 0x00, 0x00, 0x00, // pbTargetSD conformant max_count = 0
		0x00, 0x00, 0x02, 0x00, // pRootKeyID unique referent id (non-zero)
		0x60, 0x59, 0x78, 0xb9, 0x4f, 0x52, 0xdf, 0x11, // GUID Data1/2/3 little-endian ...
		0x8b, 0x6d, 0x83, 0xdc, 0xde, 0xd7, 0x20, 0x85, // ... Data4 verbatim (16 octets total)
		0x00, 0x00, 0x00, 0x00, // L0KeyID = 0
		0x01, 0x00, 0x00, 0x00, // L1KeyID = 1
		0x02, 0x00, 0x00, 0x00, // L2KeyID = 2
	}
	if !bytes.Equal(cap.stub, want) {
		t.Fatalf("request stub = % x, want % x", cap.stub, want)
	}
	// The referent id must be non-zero for a non-NULL pointer.
	if bytes.Equal(cap.stub[8:12], []byte{0, 0, 0, 0}) {
		t.Fatalf("pRootKeyID referent id is zero for a non-NULL pointer")
	}
}

// TestGetKeyResponseUnmarshal drives GetKey with a canned response stub and checks that
// ppbOut is decoded as a single conformant byte buffer behind one referent id (not a
// per-byte pointer array), that pcbOut is read inline before it, and that the trailing
// HRESULT return value maps to success / error.
func TestGetKeyResponseUnmarshal(t *testing.T) {
	// pcbOut = 4; ppbOut = referent id + max_count(4) + 4 octets; status(retval) = 0.
	respOK := []byte{
		0x04, 0x00, 0x00, 0x00, // pcbOut = 4
		0x01, 0x00, 0x02, 0x00, // non-zero referent id for the unique ppbOut pointer
		0x04, 0x00, 0x00, 0x00, // max_count = 4
		0x0a, 0x0b, 0x0c, 0x0d, // the 4 key-BLOB octets
		0x00, 0x00, 0x00, 0x00, // status = S_OK (encoded after the deferred referent)
	}
	cap := &captureInvoker{resp: respOK}
	cb, out, err := GetKey(cap, 0, []byte{}, nil, -1, -1, -1)
	if err != nil {
		t.Fatalf("GetKey returned error: %v", err)
	}
	if cb != 4 {
		t.Fatalf("pcbOut = %d, want 4", cb)
	}
	if !bytes.Equal(out, []byte{0x0a, 0x0b, 0x0c, 0x0d}) {
		t.Fatalf("ppbOut = % x, want 0a 0b 0c 0d", out)
	}

	// A nonzero HRESULT must surface as an error.
	respErr := []byte{
		0x00, 0x00, 0x00, 0x00, // pcbOut = 0
		0x00, 0x00, 0x00, 0x00, // NULL ppbOut referent
		0x05, 0x00, 0x07, 0x80, // status = E_ACCESSDENIED (0x80070005)
	}
	cap = &captureInvoker{resp: respErr}
	if _, _, err = GetKey(cap, 0, []byte{}, nil, -1, -1, -1); err == nil {
		t.Fatalf("GetKey with E_ACCESSDENIED status returned nil error")
	}
}
