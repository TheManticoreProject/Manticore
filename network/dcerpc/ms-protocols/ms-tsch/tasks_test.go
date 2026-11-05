package ms_tsch

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
	"unicode/utf16"

	schrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/86d35949-83c9-4044-b424-db363231fd0c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	mstsch "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsch"
)

// fakeTransport is an in-memory transport.Transport for driving the DCE/RPC client
// without a network.
type fakeTransport struct {
	sent      [][]byte
	recvQueue [][]byte
}

func (f *fakeTransport) Connect() error { return nil }
func (f *fakeTransport) Send(p []byte) error {
	f.sent = append(f.sent, append([]byte(nil), p...))
	return nil
}
func (f *fakeTransport) Recv() ([]byte, error) {
	if len(f.recvQueue) == 0 {
		return nil, errors.New("recv queue empty")
	}
	c := f.recvQueue[0]
	f.recvQueue = f.recvQueue[1:]
	return c, nil
}
func (f *fakeTransport) Close() error        { return nil }
func (f *fakeTransport) MaxXmitFrag() uint16 { return 4280 }
func (f *fakeTransport) MaxRecvFrag() uint16 { return 4280 }
func (f *fakeTransport) queue(b []byte)      { f.recvQueue = append(f.recvQueue, b) }

func bindAck(t *testing.T) []byte {
	t.Helper()
	ack := &pdu.BindAck{
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		Results:     []pdu.PresentationResult{{Result: pdu.ResultAcceptance, TransferSyntax: syntax.NDRTransferSyntax()}},
	}
	b, err := ack.Marshal()
	if err != nil {
		t.Fatalf("bind_ack marshal: %v", err)
	}
	return b
}

func responsePDU(t *testing.T, callID uint32, stubBytes []byte) []byte {
	t.Helper()
	resp := &pdu.Response{Stub: stubBytes}
	resp.Header = pdu.NewHeader(pdu.PacketTypeResponse, pdu.PFCFirstFrag|pdu.PFCLastFrag, callID)
	b, err := resp.Marshal()
	if err != nil {
		t.Fatalf("response marshal: %v", err)
	}
	return b
}

func stub(t *testing.T, v any) []byte {
	t.Helper()
	b, err := ndr.Marshal(v)
	if err != nil {
		t.Fatalf("marshal response stub: %v", err)
	}
	return b
}

// connectedClient returns an ms_tsch.Client whose atsvc association is already bound over
// ft, bypassing the pipe binder so the in-memory transport can drive it.
func connectedClient(t *testing.T, ft *fakeTransport) *Client {
	t.Helper()
	ft.queue(bindAck(t))
	rpc := dcerpcclient.NewClient(ft)
	if err := rpc.Bind(schrpc.SyntaxID()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	return &Client{rpc: rpc, connected: true}
}

// Mirrors of the ITaskSchedulerService [out] response structures.
type registerResp struct {
	PActualPath *ndr.WSTR                   `ndr:"unique"`
	PErrorInfo  *mstsch.TASK_XML_ERROR_INFO `ndr:"unique"`
	Status      ndr.DWORD                   `ndr:"retval"`
}

type runResp struct {
	PGuid  guid.GUID
	Status ndr.DWORD `ndr:"retval"`
}

type deleteResp struct {
	Status ndr.DWORD `ndr:"retval"`
}

func opnumOf(t *testing.T, requestPDU []byte) uint16 {
	t.Helper()
	var req pdu.Request
	if _, err := req.Unmarshal(requestPDU); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	return req.Opnum
}

// utf16le returns the UTF-16LE encoding of s, as it appears in an ndr.WSTR on the wire.
func utf16le(s string) []byte {
	units := utf16.Encode([]rune(s))
	b := make([]byte, 2*len(units))
	for i, u := range units {
		binary.LittleEndian.PutUint16(b[2*i:], u)
	}
	return b
}

// TestRegisterRunDelete drives the full atexec-style chain and checks the opnums on the
// wire, that the command is carried in the registered task XML, and that the task is
// deleted afterwards.
func TestRegisterRunDelete(t *testing.T) {
	ft := &fakeTransport{}
	c := connectedClient(t, ft)

	ft.queue(responsePDU(t, 2, stub(t, &registerResp{Status: 0}))) // SchRpcRegisterTask
	ft.queue(responsePDU(t, 3, stub(t, &runResp{Status: 0})))      // SchRpcRun
	ft.queue(responsePDU(t, 4, stub(t, &deleteResp{Status: 0})))   // SchRpcDelete

	if err := c.RegisterRunDelete("Updater", `C:\Windows\System32\cmd.exe`, `/c whoami > \\127.0.0.1\C$\o.txt`); err != nil {
		t.Fatalf("RegisterRunDelete() error = %v", err)
	}

	// Register (1), Run (12), Delete (13), in order.
	if op := opnumOf(t, ft.sent[1]); op != schrpc.OpnumSchRpcRegisterTask {
		t.Fatalf("first call opnum = %d, want SchRpcRegisterTask (%d)", op, schrpc.OpnumSchRpcRegisterTask)
	}
	if op := opnumOf(t, ft.sent[2]); op != schrpc.OpnumSchRpcRun {
		t.Fatalf("second call opnum = %d, want SchRpcRun (%d)", op, schrpc.OpnumSchRpcRun)
	}
	if op := opnumOf(t, ft.sent[3]); op != schrpc.OpnumSchRpcDelete {
		t.Fatalf("the task was not deleted: third call opnum = %d, want SchRpcDelete (%d)", op, schrpc.OpnumSchRpcDelete)
	}

	// The command must be carried in the registered XML (UTF-16LE on the wire).
	if !bytes.Contains(ft.sent[1], utf16le("cmd.exe")) {
		t.Fatal("the registered task XML does not carry the command")
	}
}

// TestRegisterFailureStops checks a failed registration is returned and the task is not
// run.
func TestRegisterFailureStops(t *testing.T) {
	ft := &fakeTransport{}
	c := connectedClient(t, ft)

	const errAccessDenied = 0x80070005 // E_ACCESSDENIED
	ft.queue(responsePDU(t, 2, stub(t, &registerResp{Status: ndr.DWORD(errAccessDenied)})))

	if err := c.RegisterRunDelete("Updater", "cmd.exe", "/c whoami"); err == nil {
		t.Fatal("RegisterRunDelete() with a failed registration: error = nil, want non-nil")
	}
	// Only the register call should have gone out (plus the bind): no run, no delete.
	if len(ft.sent) != 2 {
		t.Fatalf("after a failed registration %d calls were sent, want 2 (bind + register)", len(ft.sent))
	}
}
