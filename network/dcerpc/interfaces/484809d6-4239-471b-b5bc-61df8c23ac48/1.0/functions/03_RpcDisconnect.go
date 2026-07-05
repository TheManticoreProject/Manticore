package functions

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcDisconnectRequest carries the [in] parameters of RpcDisconnect.
type rpcDisconnectRequest struct {
	HSession mststs.SESSION_HANDLE
}

func (*rpcDisconnectRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcDisconnect }

// rpcDisconnectResponse carries the [out] parameters and return value of RpcDisconnect.
type rpcDisconnectResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDisconnect calls RpcDisconnect (opnum 3) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcDisconnect(rpc ndr.Invoker, hSession mststs.SESSION_HANDLE) (err error) {
	req := &rpcDisconnectRequest{
		HSession: hSession,
	}
	var resp rpcDisconnectResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDisconnect: %w", err)
		return
	}
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcDisconnect failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
