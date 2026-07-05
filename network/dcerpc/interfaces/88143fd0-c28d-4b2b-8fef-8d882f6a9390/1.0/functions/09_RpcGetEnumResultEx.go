package functions

import (
	"fmt"

	TermSrvEnumeration "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/88143fd0-c28d-4b2b-8fef-8d882f6a9390/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetEnumResultExRequest carries the [in] parameters of RpcGetEnumResultEx.
type rpcGetEnumResultExRequest struct {
	HEnum mststs.ENUM_HANDLE
	Level ndr.DWORD
}

func (*rpcGetEnumResultExRequest) Opnum() uint16 { return TermSrvEnumeration.OpnumRpcGetEnumResultEx }

// rpcGetEnumResultExResponse carries the [out] parameters and return value of RpcGetEnumResultEx.
type rpcGetEnumResultExResponse struct {
	PpSessionEnumResult []mststs.SESSIONENUM_EX `ndr:"unique,conformant"`
	PEntries            ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// RpcGetEnumResultEx calls RpcGetEnumResultEx (opnum 9) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetEnumResultEx(rpc ndr.Invoker, hEnum mststs.ENUM_HANDLE, level ndr.DWORD) (PpSessionEnumResult []mststs.SESSIONENUM_EX, PEntries ndr.DWORD, err error) {
	req := &rpcGetEnumResultExRequest{
		HEnum: hEnum,
		Level: level,
	}
	var resp rpcGetEnumResultExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetEnumResultEx: %w", err)
		return
	}
	PpSessionEnumResult = resp.PpSessionEnumResult
	PEntries = resp.PEntries
	if uint32(resp.Status) != TermSrvEnumeration.StatusSuccess {
		err = fmt.Errorf("RpcGetEnumResultEx failed: %s", TermSrvEnumeration.StatusString(uint32(resp.Status)))
	}
	return
}
