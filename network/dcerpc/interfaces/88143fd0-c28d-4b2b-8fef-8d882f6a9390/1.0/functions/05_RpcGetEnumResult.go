package functions

import (
	"fmt"

	TermSrvEnumeration "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/88143fd0-c28d-4b2b-8fef-8d882f6a9390/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetEnumResultRequest carries the [in] parameters of RpcGetEnumResult.
type rpcGetEnumResultRequest struct {
	HEnum mststs.ENUM_HANDLE
	Level ndr.DWORD
}

func (*rpcGetEnumResultRequest) Opnum() uint16 { return TermSrvEnumeration.OpnumRpcGetEnumResult }

// rpcGetEnumResultResponse carries the [out] parameters and return value of RpcGetEnumResult.
type rpcGetEnumResultResponse struct {
	PpSessionEnumResult []mststs.SESSIONENUM `ndr:"unique,conformant"`
	PEntries            ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// RpcGetEnumResult calls RpcGetEnumResult (opnum 5) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetEnumResult(rpc ndr.Invoker, hEnum mststs.ENUM_HANDLE, level ndr.DWORD) (PpSessionEnumResult []mststs.SESSIONENUM, PEntries ndr.DWORD, err error) {
	req := &rpcGetEnumResultRequest{
		HEnum: hEnum,
		Level: level,
	}
	var resp rpcGetEnumResultResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetEnumResult: %w", err)
		return
	}
	PpSessionEnumResult = resp.PpSessionEnumResult
	PEntries = resp.PEntries
	if uint32(resp.Status) != TermSrvEnumeration.StatusSuccess {
		err = fmt.Errorf("RpcGetEnumResult failed: %s", TermSrvEnumeration.StatusString(uint32(resp.Status)))
	}
	return
}
