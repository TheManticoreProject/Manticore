package server_test

import (
	"strings"
	"testing"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	srvsvcfunctions "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	rpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	smb2pipe "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/smb2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// TestIPCNamedPipeShareEnum checks that a CREATE of srvsvc on IPC$ opens a pipe
// and that a DCE/RPC bind plus NetrShareEnum carried over FSCTL_PIPE_TRANSCEIVE
// and READ report the server's shares, which is how smbclient -L and net view
// list shares. The server used to answer the CREATE with
// STATUS_OBJECT_NAME_NOT_FOUND and the transceive with STATUS_NOT_SUPPORTED
// (#2045).
func TestIPCNamedPipeShareEnum(t *testing.T) {
	for _, dialect := range []dialects.Dialect{dialects.SMB2_DIALECT_2_1_0, dialects.SMB2_DIALECT_3_1_1} {
		t.Run(dialect.String(), func(t *testing.T) {
			_, port := startServer(t, server.Config{})
			c, err := dial(t, port, dialect, "Passw0rd!")
			if err != nil {
				t.Fatal(err)
			}
			if err := c.TreeConnect("IPC$"); err != nil {
				t.Fatalf("TreeConnect IPC$: %v", err)
			}

			pipe := smb2pipe.New(c, `\PIPE\srvsvc`)
			if err := pipe.Connect(); err != nil {
				t.Fatalf("opening srvsvc on IPC$: %v", err)
			}
			rpc := rpcclient.NewClient(pipe)
			if err := rpc.Bind(srvsvc.SyntaxID()); err != nil {
				t.Fatalf("binding srvsvc: %v", err)
			}

			info, total, _, err := srvsvcfunctions.NetrShareEnum(rpc, "",
				mssrvs.SHARE_ENUM_STRUCT{Level: 1, ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: 1}},
				0xFFFFFFFF, nil)
			if err != nil {
				t.Fatalf("NetrShareEnum: %v", err)
			}
			if info.ShareInfo.Level1 == nil {
				t.Fatal("NetrShareEnum returned no level 1 container")
			}
			got := map[string]uint32{}
			for _, entry := range info.ShareInfo.Level1.Buffer {
				got[strings.ToUpper(string(entry.Shi1Netname))] = uint32(entry.Shi1Type)
			}
			if total != 2 || len(got) != 2 {
				t.Errorf("NetrShareEnum reported %d shares (%v), want files and IPC$", total, got)
			}
			if typ, ok := got["FILES"]; !ok || typ != 0 {
				t.Errorf("share files missing or of type 0x%08X, want STYPE_DISKTREE", typ)
			}
			if typ, ok := got["IPC$"]; !ok || typ != 0x80000003 {
				t.Errorf("share IPC$ missing or of type 0x%08X, want STYPE_IPC|STYPE_SPECIAL", typ)
			}

			if err := pipe.Close(); err != nil {
				t.Errorf("closing the pipe: %v", err)
			}
		})
	}
}

// TestIPCUnknownPipeNotFound checks that a pipe the server does not serve is
// still refused with STATUS_OBJECT_NAME_NOT_FOUND.
func TestIPCUnknownPipeNotFound(t *testing.T) {
	_, port := startServer(t, server.Config{})
	c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.TreeConnect("IPC$"); err != nil {
		t.Fatalf("TreeConnect IPC$: %v", err)
	}
	_, err = c.CreateFile("nosuchpipe", fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE, fileflags.FILE_OPEN, 0)
	if err == nil || !strings.Contains(err.Error(), "0xc0000034") {
		t.Fatalf("CREATE of an unknown pipe returned %v, want STATUS_OBJECT_NAME_NOT_FOUND", err)
	}
}

// TestIPCNamedPipeWriteRead checks the other way a client drives a pipe, used by
// impacket: the PDU is written with WRITE and the answer collected with READ,
// and a read shorter than the answer gets STATUS_BUFFER_OVERFLOW with the first
// part, then the rest.
func TestIPCNamedPipeWriteRead(t *testing.T) {
	_, port := startServer(t, server.Config{})
	c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.TreeConnect("IPC$"); err != nil {
		t.Fatalf("TreeConnect IPC$: %v", err)
	}
	id, err := c.CreateFile("srvsvc", fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE, fileflags.FILE_OPEN, 0)
	if err != nil {
		t.Fatalf("CREATE srvsvc: %v", err)
	}

	bind := &pdu.Bind{
		Header:      pdu.NewHeader(pdu.PacketTypeBind, pdu.PFCFirstFrag|pdu.PFCLastFrag, 1),
		MaxXmitFrag: 4280,
		MaxRecvFrag: 4280,
		ContextList: []pdu.ContextElement{{
			AbstractSyntax:   srvsvc.SyntaxID(),
			TransferSyntaxes: []syntax.SyntaxID{syntax.NDRTransferSyntax()},
		}},
	}
	request, err := bind.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if n, err := c.WriteFile(id, 0, request); err != nil || int(n) != len(request) {
		t.Fatalf("WRITE of the bind: %d bytes, %v", n, err)
	}
	first, err := c.ReadFile(id, 0, 16)
	if err != nil {
		t.Fatalf("first READ: %v", err)
	}
	rest, err := c.ReadFile(id, 0, 4280)
	if err != nil {
		t.Fatalf("second READ: %v", err)
	}
	reply := append(first, rest...)
	header, err := pdu.PeekHeader(reply)
	if err != nil {
		t.Fatalf("the answer is not a PDU: %v", err)
	}
	if header.PacketType != pdu.PacketTypeBindAck || int(header.FragLength) != len(reply) || len(first) != 16 {
		t.Fatalf("got %s of %d bytes (frag_length %d, first read %d), want a whole bind_ack split 16+rest",
			header.PacketType, len(reply), header.FragLength, len(first))
	}
	if err := c.CloseFile(id); err != nil {
		t.Errorf("CLOSE: %v", err)
	}
}
