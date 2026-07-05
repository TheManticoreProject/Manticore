package mssrvs

import (
	"testing"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
	srvstypes "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// fakePipeDialer is a PipeDialer that hands out a pre-scripted in-memory transport,
// recording the pipe name it was asked to open.
type fakePipeDialer struct {
	ft        *fakeTransport
	lastPipe  string
	openCalls int
}

func (d *fakePipeDialer) RPCTransport(pipeName string) (dcerpctransport.Transport, error) {
	d.lastPipe = pipeName
	d.openCalls++
	return d.ft, nil
}

// TestNewUsesDialerToBindSrvsvc verifies the decoupled path: New takes a PipeDialer,
// and a workflow opens the \srvsvc pipe through it, binds, and issues the call — all
// without any concrete SMB-client dependency.
func TestNewUsesDialerToBindSrvsvc(t *testing.T) {
	ft := &fakeTransport{}
	ft.queue(bindAck(t)) // the Bind in c.bind() consumes this

	// One share in a level-1 container, then the NetrShareEnum reply.
	resp := &shareEnumResp{
		InfoStruct: srvstypes.SHARE_ENUM_STRUCT{
			Level: 1,
			ShareInfo: srvstypes.SHARE_ENUM_UNION{
				Tag: 1,
				Level1: &srvstypes.SHARE_INFO_1_CONTAINER{
					EntriesRead: 1,
					Buffer:      []srvstypes.SHARE_INFO_1{{Shi1Netname: "C$", Shi1Type: 0x80000000, Shi1Remark: "Default share"}},
				},
			},
		},
		TotalEntries: 1,
		Status:       ndr.DWORD(srvsvc.NERR_Success),
	}
	ft.queue(responsePDU(t, 2, stub(t, resp)))

	dialer := &fakePipeDialer{ft: ft}
	c := New(dialer)

	shares, err := c.ListShares()
	if err != nil {
		t.Fatalf("ListShares() error = %v", err)
	}
	if len(shares) != 1 || shares[0].Name != "C$" {
		t.Fatalf("shares = %+v, want one entry C$", shares)
	}
	if dialer.openCalls != 1 {
		t.Errorf("dialer opened %d pipes, want 1", dialer.openCalls)
	}
	if dialer.lastPipe != srvsvc.PipeName {
		t.Errorf("opened pipe %q, want %q", dialer.lastPipe, srvsvc.PipeName)
	}
}
