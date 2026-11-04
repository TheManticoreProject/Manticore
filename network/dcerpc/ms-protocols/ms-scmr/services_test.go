package ms_scmr

import (
	"errors"
	"testing"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
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

// connectedClient returns an ms_scmr.Client whose svcctl association is already bound over
// ft, bypassing the pipe binder so the in-memory transport can drive it.
func connectedClient(t *testing.T, ft *fakeTransport) *Client {
	t.Helper()
	ft.queue(bindAck(t))
	rpc := dcerpcclient.NewClient(ft)
	if err := rpc.Bind(svcctl.SyntaxID()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	return &Client{rpc: rpc, connected: true}
}

// Mirrors of the svcctl [out] response structures, for marshalling canned replies.
type openSCMResp struct {
	Handle msscmr.SC_RPC_HANDLE
	Status ndr.DWORD `ndr:"retval"`
}

type createSvcResp struct {
	TagID  *ndr.DWORD `ndr:"unique"`
	Handle msscmr.SC_RPC_HANDLE
	Status ndr.DWORD `ndr:"retval"`
}

type statusResp struct {
	Status ndr.DWORD `ndr:"retval"`
}

type closeResp struct {
	Handle msscmr.SC_RPC_HANDLE
	Status ndr.DWORD `ndr:"retval"`
}

func sentinelHandle() msscmr.SC_RPC_HANDLE {
	var h msscmr.SC_RPC_HANDLE
	for i := range h {
		h[i] = byte(i + 1)
	}
	return h
}

// opnumOf parses a request PDU and returns its opnum.
func opnumOf(t *testing.T, requestPDU []byte) uint16 {
	t.Helper()
	var req pdu.Request
	if _, err := req.Unmarshal(requestPDU); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	return req.Opnum
}

// TestOpenSCManager round-trips a canned handle and checks the opnum on the wire.
func TestOpenSCManager(t *testing.T) {
	ft := &fakeTransport{}
	c := connectedClient(t, ft)

	want := sentinelHandle()
	ft.queue(responsePDU(t, 2, stub(t, &openSCMResp{Handle: want, Status: 0})))

	got, err := c.OpenSCManager(SCManagerCreateService)
	if err != nil {
		t.Fatalf("OpenSCManager() error = %v", err)
	}
	if got != want {
		t.Fatalf("OpenSCManager() handle = %x, want %x", got, want)
	}
	if op := opnumOf(t, ft.sent[1]); op != svcctl.OpnumROpenSCManagerW {
		t.Fatalf("opnum = %d, want %d", op, svcctl.OpnumROpenSCManagerW)
	}
}

// TestCreateAndRunToleratesStartError drives the full service-exec chain. The service's
// start returns a non-zero Win32 status (as it always does for a plain command, which
// never checks in as a service); CreateAndRun must treat that as the command having run,
// still delete the service, and return nil.
func TestCreateAndRunToleratesStartError(t *testing.T) {
	ft := &fakeTransport{}
	c := connectedClient(t, ft)

	const errorServiceRequestTimeout = 1053 // ERROR_SERVICE_REQUEST_TIMEOUT
	scm := sentinelHandle()
	svc := sentinelHandle()
	ft.queue(responsePDU(t, 2, stub(t, &openSCMResp{Handle: scm})))                       // OpenSCManager
	ft.queue(responsePDU(t, 3, stub(t, &createSvcResp{Handle: svc})))                     // CreateService
	ft.queue(responsePDU(t, 4, stub(t, &statusResp{Status: errorServiceRequestTimeout}))) // StartService (expected failure)
	ft.queue(responsePDU(t, 5, stub(t, &statusResp{Status: 0})))                          // DeleteService
	ft.queue(responsePDU(t, 6, stub(t, &closeResp{})))                                    // CloseHandle(svc)
	ft.queue(responsePDU(t, 7, stub(t, &closeResp{})))                                    // CloseHandle(scm)

	if err := c.CreateAndRun("Updater", `C:\Windows\System32\cmd.exe /c whoami > \\127.0.0.1\C$\o.txt`); err != nil {
		t.Fatalf("CreateAndRun() error = %v, want nil (the start error is expected)", err)
	}

	// The exec chain must create (12), start (19) and, crucially, delete (2) the service.
	if op := opnumOf(t, ft.sent[2]); op != svcctl.OpnumRCreateServiceW {
		t.Fatalf("second call opnum = %d, want RCreateServiceW (%d)", op, svcctl.OpnumRCreateServiceW)
	}
	if op := opnumOf(t, ft.sent[4]); op != svcctl.OpnumRDeleteService {
		t.Fatalf("the service was not deleted: fourth call opnum = %d, want RDeleteService (%d)", op, svcctl.OpnumRDeleteService)
	}
}
