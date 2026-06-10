package mssrvs

import (
	"testing"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

type serverGetInfoResp struct {
	InfoStruct structures.SERVER_INFO
	Status     ndr.DWORD `ndr:"retval"`
}

func TestGetServerInfo(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)

	resp := &serverGetInfoResp{
		InfoStruct: structures.SERVER_INFO{
			Tag: 101,
			ServerInfo101: &structures.SERVER_INFO_101{
				Sv101PlatformId:   500,
				Sv101Name:         "FILESERVER",
				Sv101VersionMajor: 6,
				Sv101VersionMinor: 1,
				Sv101Type:         0x00000003,
				Sv101Comment:      "main file server",
			},
		},
		Status: ndr.DWORD(srvsvc.NERR_Success),
	}
	ft.queue(responsePDU(t, 2, stub(t, resp)))

	info, err := getServerInfo(c)
	if err != nil {
		t.Fatalf("getServerInfo() error = %v", err)
	}
	if info == nil {
		t.Fatal("getServerInfo() returned nil info")
	}
	if info.PlatformID != 500 || info.Name != "FILESERVER" || info.VersionMajor != 6 ||
		info.VersionMinor != 1 || info.Type != 0x00000003 || info.Comment != "main file server" {
		t.Errorf("server info = %+v", info)
	}

	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	if req.Opnum != srvsvc.OpnumNetrServerGetInfo {
		t.Errorf("opnum = %d, want %d", req.Opnum, srvsvc.OpnumNetrServerGetInfo)
	}
}

func TestGetServerInfo_MissingArm(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)
	// Success status but a nil level-101 arm must be reported, not nil-dereferenced.
	resp := &serverGetInfoResp{
		InfoStruct: structures.SERVER_INFO{Tag: 101},
		Status:     ndr.DWORD(srvsvc.NERR_Success),
	}
	ft.queue(responsePDU(t, 2, stub(t, resp)))
	if _, err := getServerInfo(c); err == nil {
		t.Fatal("getServerInfo() with missing arm: error = nil, want non-nil")
	}
}
