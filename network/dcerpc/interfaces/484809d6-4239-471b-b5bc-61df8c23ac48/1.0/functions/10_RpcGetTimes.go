package functions

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetTimesRequest carries the [in] parameters of RpcGetTimes.
type rpcGetTimesRequest struct {
	HSession mststs.SESSION_HANDLE
}

func (*rpcGetTimesRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcGetTimes }

// rpcGetTimesResponse carries the [out] parameters and return value of RpcGetTimes.
type rpcGetTimesResponse struct {
	PConnectTime    int64
	PDisconnectTime int64
	PLogonTime      int64
	Status          ndr.DWORD `ndr:"retval"`
}

// RpcGetTimes calls RpcGetTimes (opnum 10) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetTimes(rpc ndr.Invoker, hSession mststs.SESSION_HANDLE) (PConnectTime int64, PDisconnectTime int64, PLogonTime int64, err error) {
	req := &rpcGetTimesRequest{
		HSession: hSession,
	}
	var resp rpcGetTimesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetTimes: %w", err)
		return
	}
	PConnectTime = resp.PConnectTime
	PDisconnectTime = resp.PDisconnectTime
	PLogonTime = resp.PLogonTime
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcGetTimes failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
