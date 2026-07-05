package mssrvs

import (
	"testing"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	srvstypes "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

type sessionEnumResp struct {
	InfoStruct   srvstypes.SESSION_ENUM_STRUCT
	TotalEntries ndr.DWORD
	ResumeHandle *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

func TestListSessions(t *testing.T) {
	ft := &fakeTransport{}
	c := boundClient(t, ft)

	resp := &sessionEnumResp{
		InfoStruct: srvstypes.SESSION_ENUM_STRUCT{
			Level: 10,
			SessionInfo: srvstypes.SESSION_ENUM_UNION{
				Tag: 10,
				Level10: &srvstypes.SESSION_INFO_10_CONTAINER{
					EntriesRead: 1,
					Buffer: []srvstypes.SESSION_INFO_10{
						{Sesi10Cname: "\\\\10.0.0.5", Sesi10Username: "ADMIN", Sesi10Time: 3600, Sesi10IdleTime: 12},
					},
				},
			},
		},
		TotalEntries: 1,
		Status:       ndr.DWORD(srvsvc.NERR_Success),
	}
	ft.queue(responsePDU(t, 2, stub(t, resp)))

	sessions, err := listSessions(c)
	if err != nil {
		t.Fatalf("listSessions() error = %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("got %d sessions, want 1", len(sessions))
	}
	got := sessions[0]
	if got.ClientName != "\\\\10.0.0.5" || got.UserName != "ADMIN" || got.ActiveSecs != 3600 || got.IdleSecs != 12 {
		t.Errorf("session[0] = %+v", got)
	}

	var req pdu.Request
	if _, err := req.Unmarshal(ft.sent[1]); err != nil {
		t.Fatalf("request does not parse: %v", err)
	}
	if req.Opnum != srvsvc.OpnumNetrSessionEnum {
		t.Errorf("opnum = %d, want %d", req.Opnum, srvsvc.OpnumNetrSessionEnum)
	}
}
