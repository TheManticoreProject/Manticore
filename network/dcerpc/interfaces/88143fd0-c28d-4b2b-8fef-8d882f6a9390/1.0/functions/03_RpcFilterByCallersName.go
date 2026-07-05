package functions

import (
	"fmt"

	TermSrvEnumeration "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/88143fd0-c28d-4b2b-8fef-8d882f6a9390/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcFilterByCallersNameRequest carries the [in] parameters of RpcFilterByCallersName.
type rpcFilterByCallersNameRequest struct {
	HEnum mststs.ENUM_HANDLE
}

func (*rpcFilterByCallersNameRequest) Opnum() uint16 {
	return TermSrvEnumeration.OpnumRpcFilterByCallersName
}

// rpcFilterByCallersNameResponse carries the [out] parameters and return value of RpcFilterByCallersName.
type rpcFilterByCallersNameResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcFilterByCallersName calls RpcFilterByCallersName (opnum 3) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcFilterByCallersName(rpc ndr.Invoker, hEnum mststs.ENUM_HANDLE) (err error) {
	req := &rpcFilterByCallersNameRequest{
		HEnum: hEnum,
	}
	var resp rpcFilterByCallersNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcFilterByCallersName: %w", err)
		return
	}
	if uint32(resp.Status) != TermSrvEnumeration.StatusSuccess {
		err = fmt.Errorf("RpcFilterByCallersName failed: %s", TermSrvEnumeration.StatusString(uint32(resp.Status)))
	}
	return
}
