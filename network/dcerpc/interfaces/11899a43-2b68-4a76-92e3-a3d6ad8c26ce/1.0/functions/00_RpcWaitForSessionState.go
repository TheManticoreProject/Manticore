package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	TermSrvNotification "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/11899a43-2b68-4a76-92e3-a3d6ad8c26ce/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcWaitForSessionStateRequest carries the [in] parameters of RpcWaitForSessionState.
type rpcWaitForSessionStateRequest struct {
	SessionId int32
	State     int32
	Timeout   ndr.DWORD
}

func (*rpcWaitForSessionStateRequest) Opnum() uint16 {
	return TermSrvNotification.OpnumRpcWaitForSessionState
}

// rpcWaitForSessionStateResponse carries the [out] parameters and return value of RpcWaitForSessionState.
type rpcWaitForSessionStateResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcWaitForSessionState calls RpcWaitForSessionState (opnum 0) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWaitForSessionState(rpc ndr.Invoker, sessionId int32, state int32, timeout ndr.DWORD) (err error) {
	req := &rpcWaitForSessionStateRequest{
		SessionId: sessionId,
		State:     state,
		Timeout:   timeout,
	}
	var resp rpcWaitForSessionStateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWaitForSessionState: %w", err)
		return
	}
	if uint32(resp.Status) != TermSrvNotification.StatusSuccess {
		err = fmt.Errorf("RpcWaitForSessionState failed: %s", TermSrvNotification.StatusString(uint32(resp.Status)))
	}
	return
}
