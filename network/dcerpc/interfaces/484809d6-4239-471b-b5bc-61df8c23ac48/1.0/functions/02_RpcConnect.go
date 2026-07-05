package functions

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcConnectRequest carries the [in] parameters of RpcConnect.
type rpcConnectRequest struct {
	HSession        mststs.SESSION_HANDLE
	TargetSessionId int32
	SzPassword      ndr.WSTR
}

func (*rpcConnectRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcConnect }

// rpcConnectResponse carries the [out] parameters and return value of RpcConnect.
type rpcConnectResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcConnect calls RpcConnect (opnum 2) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcConnect(rpc ndr.Invoker, hSession mststs.SESSION_HANDLE, targetSessionId int32, szPassword ndr.WSTR) (err error) {
	req := &rpcConnectRequest{
		HSession:        hSession,
		TargetSessionId: targetSessionId,
		SzPassword:      szPassword,
	}
	var resp rpcConnectResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcConnect: %w", err)
		return
	}
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcConnect failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
