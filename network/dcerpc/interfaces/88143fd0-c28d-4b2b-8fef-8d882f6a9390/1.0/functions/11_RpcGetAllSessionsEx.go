package functions

import (
	"fmt"

	TermSrvEnumeration "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/88143fd0-c28d-4b2b-8fef-8d882f6a9390/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetAllSessionsExRequest carries the [in] parameters of RpcGetAllSessionsEx.
type rpcGetAllSessionsExRequest struct {
	Level ndr.DWORD
}

func (*rpcGetAllSessionsExRequest) Opnum() uint16 { return TermSrvEnumeration.OpnumRpcGetAllSessionsEx }

// rpcGetAllSessionsExResponse carries the [out] parameters and return value of RpcGetAllSessionsEx.
type rpcGetAllSessionsExResponse struct {
	PpSessionData []mststs.EXECENVDATAEX `ndr:"unique,conformant"`
	PcEntries     ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// RpcGetAllSessionsEx calls RpcGetAllSessionsEx (opnum 11) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetAllSessionsEx(rpc ndr.Invoker, level ndr.DWORD) (PpSessionData []mststs.EXECENVDATAEX, PcEntries ndr.DWORD, err error) {
	req := &rpcGetAllSessionsExRequest{
		Level: level,
	}
	var resp rpcGetAllSessionsExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetAllSessionsEx: %w", err)
		return
	}
	PpSessionData = resp.PpSessionData
	PcEntries = resp.PcEntries
	if uint32(resp.Status) != TermSrvEnumeration.StatusSuccess {
		err = fmt.Errorf("RpcGetAllSessionsEx failed: %s", TermSrvEnumeration.StatusString(uint32(resp.Status)))
	}
	return
}
