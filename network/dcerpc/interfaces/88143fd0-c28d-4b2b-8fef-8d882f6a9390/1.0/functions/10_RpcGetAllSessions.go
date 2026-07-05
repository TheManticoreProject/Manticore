package functions

import (
	"fmt"

	TermSrvEnumeration "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/88143fd0-c28d-4b2b-8fef-8d882f6a9390/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetAllSessionsRequest carries the [in] parameters of RpcGetAllSessions.
type rpcGetAllSessionsRequest struct {
	PLevel ndr.DWORD
}

func (*rpcGetAllSessionsRequest) Opnum() uint16 { return TermSrvEnumeration.OpnumRpcGetAllSessions }

// rpcGetAllSessionsResponse carries the [out] parameters and return value of RpcGetAllSessions.
type rpcGetAllSessionsResponse struct {
	PLevel        ndr.DWORD
	PpSessionData []mststs.EXECENVDATA `ndr:"unique,conformant"`
	PcEntries     ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// RpcGetAllSessions calls RpcGetAllSessions (opnum 10) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetAllSessions(rpc ndr.Invoker, pLevel ndr.DWORD) (PLevel ndr.DWORD, PpSessionData []mststs.EXECENVDATA, PcEntries ndr.DWORD, err error) {
	req := &rpcGetAllSessionsRequest{
		PLevel: pLevel,
	}
	var resp rpcGetAllSessionsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetAllSessions: %w", err)
		return
	}
	PLevel = resp.PLevel
	PpSessionData = resp.PpSessionData
	PcEntries = resp.PcEntries
	if uint32(resp.Status) != TermSrvEnumeration.StatusSuccess {
		err = fmt.Errorf("RpcGetAllSessions failed: %s", TermSrvEnumeration.StatusString(uint32(resp.Status)))
	}
	return
}
