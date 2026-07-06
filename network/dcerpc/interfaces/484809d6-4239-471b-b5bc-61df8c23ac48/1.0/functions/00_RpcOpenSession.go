package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcOpenSessionRequest carries the [in] parameters of RpcOpenSession.
type rpcOpenSessionRequest struct {
	SessionId int32
}

func (*rpcOpenSessionRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcOpenSession }

// rpcOpenSessionResponse carries the [out] parameters and return value of RpcOpenSession.
type rpcOpenSessionResponse struct {
	PhSession mststs.SESSION_HANDLE
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcOpenSession calls RpcOpenSession (opnum 0) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcOpenSession(rpc ndr.Invoker, sessionId int32) (PhSession mststs.SESSION_HANDLE, err error) {
	req := &rpcOpenSessionRequest{
		SessionId: sessionId,
	}
	var resp rpcOpenSessionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcOpenSession: %w", err)
		return
	}
	PhSession = resp.PhSession
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcOpenSession failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
