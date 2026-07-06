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

// rpcLogoffRequest carries the [in] parameters of RpcLogoff.
type rpcLogoffRequest struct {
	HSession mststs.SESSION_HANDLE
}

func (*rpcLogoffRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcLogoff }

// rpcLogoffResponse carries the [out] parameters and return value of RpcLogoff.
type rpcLogoffResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcLogoff calls RpcLogoff (opnum 4) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcLogoff(rpc ndr.Invoker, hSession mststs.SESSION_HANDLE) (err error) {
	req := &rpcLogoffRequest{
		HSession: hSession,
	}
	var resp rpcLogoffResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcLogoff: %w", err)
		return
	}
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcLogoff failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
