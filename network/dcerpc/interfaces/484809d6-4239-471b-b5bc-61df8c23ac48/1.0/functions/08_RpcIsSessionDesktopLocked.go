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

// rpcIsSessionDesktopLockedRequest carries the [in] parameters of RpcIsSessionDesktopLocked.
type rpcIsSessionDesktopLockedRequest struct {
	HSession mststs.SESSION_HANDLE
}

func (*rpcIsSessionDesktopLockedRequest) Opnum() uint16 {
	return TermSrvSession.OpnumRpcIsSessionDesktopLocked
}

// rpcIsSessionDesktopLockedResponse carries the [out] parameters and return value of RpcIsSessionDesktopLocked.
type rpcIsSessionDesktopLockedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcIsSessionDesktopLocked calls RpcIsSessionDesktopLocked (opnum 8) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcIsSessionDesktopLocked(rpc ndr.Invoker, hSession mststs.SESSION_HANDLE) (err error) {
	req := &rpcIsSessionDesktopLockedRequest{
		HSession: hSession,
	}
	var resp rpcIsSessionDesktopLockedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcIsSessionDesktopLocked: %w", err)
		return
	}
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcIsSessionDesktopLocked failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
