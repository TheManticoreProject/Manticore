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

// rpcCloseSessionRequest carries the [in] parameters of RpcCloseSession.
type rpcCloseSessionRequest struct {
	PhSession mststs.SESSION_HANDLE
}

func (*rpcCloseSessionRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcCloseSession }

// rpcCloseSessionResponse carries the [out] parameters and return value of RpcCloseSession.
type rpcCloseSessionResponse struct {
	PhSession mststs.SESSION_HANDLE
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcCloseSession calls RpcCloseSession (opnum 1) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcCloseSession(rpc ndr.Invoker, phSession mststs.SESSION_HANDLE) (PhSession mststs.SESSION_HANDLE, err error) {
	req := &rpcCloseSessionRequest{
		PhSession: phSession,
	}
	var resp rpcCloseSessionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcCloseSession: %w", err)
		return
	}
	PhSession = resp.PhSession
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcCloseSession failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
