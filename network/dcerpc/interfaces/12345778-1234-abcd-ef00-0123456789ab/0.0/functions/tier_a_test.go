package functions_test

import (
	"bytes"
	"testing"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

func TestLsarDeleteObject_RoundTrip(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)
	ft.queue(responsePDU(t, 2, make([]byte, 24))) // zeroed handle + STATUS_SUCCESS

	in := structures.LSAPR_HANDLE{0x00, 0x00, 0x00, 0x00, 0xaa, 0xbb}
	out, err := functions.LsarDeleteObject(c, in)
	if err != nil {
		t.Fatalf("LsarDeleteObject() error = %v", err)
	}
	if !out.IsZero() {
		t.Errorf("handle after delete = %x, want zeroed", out)
	}
	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	if req.Opnum != lsarpc.OpnumLsarDeleteObject {
		t.Errorf("opnum = %d, want %d", req.Opnum, lsarpc.OpnumLsarDeleteObject)
	}
	if !bytes.Equal(req.Stub, in[:]) {
		t.Errorf("request stub = %x, want %x", req.Stub, in[:])
	}
}

func TestLsarGetSystemAccessAccount_RoundTrip(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)
	// [out] SystemAccess (0x00000007) + STATUS_SUCCESS.
	ft.queue(responsePDU(t, 2, []byte{0x07, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}))

	h := structures.LSAPR_HANDLE{0x00, 0x00, 0x00, 0x00, 0x11, 0x22}
	access, err := functions.LsarGetSystemAccessAccount(c, h)
	if err != nil {
		t.Fatalf("LsarGetSystemAccessAccount() error = %v", err)
	}
	if access != 0x00000007 {
		t.Errorf("system access = %#x, want 0x7", access)
	}
	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	if req.Opnum != lsarpc.OpnumLsarGetSystemAccessAccount {
		t.Errorf("opnum = %d, want %d", req.Opnum, lsarpc.OpnumLsarGetSystemAccessAccount)
	}
	if !bytes.Equal(req.Stub, h[:]) {
		t.Errorf("request stub = %x, want %x", req.Stub, h[:])
	}
}

func TestLsarSetSystemAccessAccount_RequestMarshalling(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)
	ft.queue(responsePDU(t, 2, make([]byte, 4))) // STATUS_SUCCESS

	h := structures.LSAPR_HANDLE{0x00, 0x00, 0x00, 0x00, 0x33, 0x44}
	if err := functions.LsarSetSystemAccessAccount(c, h, 0x00000003); err != nil {
		t.Fatalf("LsarSetSystemAccessAccount() error = %v", err)
	}
	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	if req.Opnum != lsarpc.OpnumLsarSetSystemAccessAccount {
		t.Errorf("opnum = %d, want %d", req.Opnum, lsarpc.OpnumLsarSetSystemAccessAccount)
	}
	// stub = 20-byte handle + DWORD SystemAccess (little-endian).
	want := append(append([]byte(nil), h[:]...), 0x03, 0x00, 0x00, 0x00)
	if !bytes.Equal(req.Stub, want) {
		t.Errorf("request stub:\n got %x\nwant %x", req.Stub, want)
	}
}

func TestLsarOpenPolicy_RequestMarshalling(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)
	wantHandle := structures.LSAPR_HANDLE{0x00, 0x00, 0x00, 0x00, 0xde, 0xad}
	respStub := append(append([]byte(nil), wantHandle[:]...), 0x00, 0x00, 0x00, 0x00)
	ft.queue(responsePDU(t, 2, respStub))

	h, err := functions.LsarOpenPolicy(c, lsarpc.MaximumAllowed)
	if err != nil {
		t.Fatalf("LsarOpenPolicy() error = %v", err)
	}
	if h != wantHandle {
		t.Errorf("handle = %x, want %x", h, wantHandle)
	}
	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	if req.Opnum != lsarpc.OpnumLsarOpenPolicy {
		t.Errorf("opnum = %d, want %d", req.Opnum, lsarpc.OpnumLsarOpenPolicy)
	}
	// NULL SystemName ptr (4) + zero ObjectAttributes (24) + DesiredAccess (4).
	want := make([]byte, 0, 32)
	want = append(want, 0, 0, 0, 0)
	want = append(want, make([]byte, 24)...)
	want = append(want, 0x00, 0x00, 0x00, 0x02) // MAXIMUM_ALLOWED 0x02000000 LE
	if !bytes.Equal(req.Stub, want) {
		t.Errorf("request stub:\n got %x\nwant %x", req.Stub, want)
	}
}
