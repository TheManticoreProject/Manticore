package mssrvs

import (
	"errors"
	"testing"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
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

// boundClient returns a DCE/RPC client already bound to srvsvc over ft.
func boundClient(t *testing.T, ft *fakeTransport) *dcerpcclient.Client {
	t.Helper()
	ft.queue(bindAck(t))
	c := dcerpcclient.NewClient(ft)
	if err := c.Bind(srvsvc.SyntaxID()); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	return c
}

// stub marshals v (a mirror of a srvsvc [out] response) to its NDR octet stream.
func stub(t *testing.T, v any) []byte {
	t.Helper()
	b, err := ndr.Marshal(v)
	if err != nil {
		t.Fatalf("marshal response stub: %v", err)
	}
	return b
}

type shareEnumResp struct {
	InfoStruct   structures.SHARE_ENUM_STRUCT
	TotalEntries ndr.DWORD
	ResumeHandle *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

func TestListShares(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)

	resp := &shareEnumResp{
		InfoStruct: structures.SHARE_ENUM_STRUCT{
			Level: 1,
			ShareInfo: structures.SHARE_ENUM_UNION{
				Tag: 1,
				Level1: &structures.SHARE_INFO_1_CONTAINER{
					EntriesRead: 2,
					Buffer: []structures.SHARE_INFO_1{
						{Shi1Netname: "C$", Shi1Type: 0x80000000, Shi1Remark: "Default share"},
						{Shi1Netname: "netlogon", Shi1Type: 0, Shi1Remark: "Logon server share"},
					},
				},
			},
		},
		TotalEntries: 2,
		Status:       ndr.DWORD(srvsvc.NERR_Success),
	}
	ft.queue(responsePDU(t, 2, stub(t, resp)))

	shares, err := listShares(c)
	if err != nil {
		t.Fatalf("listShares() error = %v", err)
	}
	if len(shares) != 2 {
		t.Fatalf("got %d shares, want 2", len(shares))
	}
	if shares[0].Name != "C$" || shares[0].Type != 0x80000000 || shares[0].Comment != "Default share" {
		t.Errorf("share[0] = %+v, want {C$ 0x80000000 Default share}", shares[0])
	}
	if shares[1].Name != "netlogon" || shares[1].Comment != "Logon server share" {
		t.Errorf("share[1] = %+v", shares[1])
	}

	// Verify the request opnum.
	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	if req.Opnum != srvsvc.OpnumNetrShareEnum {
		t.Errorf("opnum = %d, want %d", req.Opnum, srvsvc.OpnumNetrShareEnum)
	}
}

func TestListShares_AccessDenied(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)
	resp := &shareEnumResp{
		InfoStruct: structures.SHARE_ENUM_STRUCT{Level: 1, ShareInfo: structures.SHARE_ENUM_UNION{Tag: 1}},
		Status:     ndr.DWORD(srvsvc.ERROR_ACCESS_DENIED),
	}
	ft.queue(responsePDU(t, 2, stub(t, resp)))
	if _, err := listShares(c); err == nil {
		t.Fatal("listShares() with ACCESS_DENIED: error = nil, want non-nil")
	}
}
