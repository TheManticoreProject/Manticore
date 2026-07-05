package functions

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetTerminalNameRequest carries the [in] parameters of RpcGetTerminalName.
type rpcGetTerminalNameRequest struct {
	HSession mststs.SESSION_HANDLE
}

func (*rpcGetTerminalNameRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcGetTerminalName }

// rpcGetTerminalNameResponse carries the [out] parameters and return value of RpcGetTerminalName.
type rpcGetTerminalNameResponse struct {
	PszTerminalName ndr.WSTR
	Status          ndr.DWORD `ndr:"retval"`
}

// RpcGetTerminalName calls RpcGetTerminalName (opnum 6) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetTerminalName(rpc ndr.Invoker, hSession mststs.SESSION_HANDLE) (PszTerminalName ndr.WSTR, err error) {
	req := &rpcGetTerminalNameRequest{
		HSession: hSession,
	}
	var resp rpcGetTerminalNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetTerminalName: %w", err)
		return
	}
	PszTerminalName = resp.PszTerminalName
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcGetTerminalName failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
