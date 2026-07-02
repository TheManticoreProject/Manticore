package functions

import (
	"bytes"
	"testing"

	BackupKey "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/3dde7c30-165d-11d1-ab8f-00805f14db40/1.0"
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

// TestBackuprKeyRequestMarshal asserts opnum 0 and, critically, that pguidActionAgent is
// carried as the 16-octet [MS-DTYP] wire GUID (Data1/2/3 little-endian, Data4 verbatim) —
// not windows/guid.GUID's 24-octet reflected layout — followed by the conformant pDataIn
// (max_count + octets, no referent id) and the trailing cbDataIn/dwParam DWORDs.
func TestBackuprKeyRequestMarshal(t *testing.T) {
	cap := &captureInvoker{}
	_, _, _ = BackuprKey(cap, BackupKey.BackupKeyBackupGUID, []uint8{0xaa, 0xbb, 0xcc}, 3, 0)

	if cap.opnum != BackupKey.OpnumBackuprKey {
		t.Fatalf("opnum = %d, want %d", cap.opnum, BackupKey.OpnumBackuprKey)
	}
	// 7f752b10-178e-11d1-ab8f-00805f14db40 on the wire.
	wantGUID := []byte{0x10, 0x2b, 0x75, 0x7f, 0x8e, 0x17, 0xd1, 0x11, 0xab, 0x8f, 0x00, 0x80, 0x5f, 0x14, 0xdb, 0x40}
	if len(cap.stub) < 16 || !bytes.Equal(cap.stub[:16], wantGUID) {
		t.Fatalf("leading GUID = % x, want % x", cap.stub[:min(16, len(cap.stub))], wantGUID)
	}
	// The request must be 32 octets: GUID(16) + conf_count(4) + data(3) + pad(1) + cbDataIn(4) + dwParam(4).
	want := []byte{
		0x10, 0x2b, 0x75, 0x7f, 0x8e, 0x17, 0xd1, 0x11, 0xab, 0x8f, 0x00, 0x80, 0x5f, 0x14, 0xdb, 0x40,
		0x03, 0x00, 0x00, 0x00, // conformant max_count = 3 (no referent id: ref pointer)
		0xaa, 0xbb, 0xcc, 0x00, // pDataIn + 1 pad octet
		0x03, 0x00, 0x00, 0x00, // cbDataIn = 3
		0x00, 0x00, 0x00, 0x00, // dwParam = 0
	}
	if !bytes.Equal(cap.stub, want) {
		t.Fatalf("request stub = % x, want % x", cap.stub, want)
	}
}

// TestBackuprKeyResponseUnmarshal drives BackuprKey with a canned response stub and checks
// that ppDataOut is decoded as a single conformant byte buffer behind one referent id
// (PBYTE_ARRAY), not a per-byte pointer array, and that the return code maps to an error.
func TestBackuprKeyResponseUnmarshal(t *testing.T) {
	// ppDataOut = {01,02,03,04}: referent id + max_count(4) + 4 octets; pcbDataOut = 4; status = 0.
	respOK := []byte{
		0x01, 0x00, 0x02, 0x00, // non-zero referent id for the unique ppDataOut pointer
		0x04, 0x00, 0x00, 0x00, // max_count = 4
		0x01, 0x02, 0x03, 0x04, // the 4 octets
		0x04, 0x00, 0x00, 0x00, // pcbDataOut = 4
		0x00, 0x00, 0x00, 0x00, // status = ERROR_SUCCESS
	}
	// pDataIn is a top-level [in] ref pointer (no referent id), so it must be non-nil even
	// when the action ignores it — the retrieve operation sends an empty conformant array.
	cap := &captureInvoker{resp: respOK}
	out, cb, err := BackuprKey(cap, BackupKey.BackupKeyRetrieveBackupKeyGUID, []uint8{}, 0, 0)
	if err != nil {
		t.Fatalf("BackuprKey returned error: %v", err)
	}
	if cb != 4 {
		t.Fatalf("pcbDataOut = %d, want 4", cb)
	}
	if !bytes.Equal(out, []byte{0x01, 0x02, 0x03, 0x04}) {
		t.Fatalf("ppDataOut = % x, want 01 02 03 04", out)
	}

	// A nonzero status must surface as an error.
	respErr := []byte{
		0x00, 0x00, 0x00, 0x00, // NULL ppDataOut referent
		0x00, 0x00, 0x00, 0x00, // pcbDataOut = 0
		0x57, 0x00, 0x00, 0x00, // status = ERROR_INVALID_PARAMETER
	}
	cap = &captureInvoker{resp: respErr}
	if _, _, err = BackuprKey(cap, guid.GUID{}, []uint8{}, 0, 0); err == nil {
		t.Fatalf("BackuprKey with ERROR_INVALID_PARAMETER status returned nil error")
	}
}
