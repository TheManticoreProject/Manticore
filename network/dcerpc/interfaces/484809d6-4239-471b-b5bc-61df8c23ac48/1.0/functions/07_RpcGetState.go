package functions

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetStateRequest carries the [in] parameters of RpcGetState.
type rpcGetStateRequest struct {
	HSession mststs.SESSION_HANDLE
}

func (*rpcGetStateRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcGetState }

// rpcGetStateResponse carries the [out] parameters and return value of RpcGetState.
type rpcGetStateResponse struct {
	PlState int32
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcGetState calls RpcGetState (opnum 7) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetState(rpc ndr.Invoker, hSession mststs.SESSION_HANDLE) (PlState int32, err error) {
	req := &rpcGetStateRequest{
		HSession: hSession,
	}
	var resp rpcGetStateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetState: %w", err)
		return
	}
	PlState = resp.PlState
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcGetState failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
