package functions

import (
	"fmt"

	TermSrvEnumeration "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/88143fd0-c28d-4b2b-8fef-8d882f6a9390/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcFilterByStateRequest carries the [in] parameters of RpcFilterByState.
type rpcFilterByStateRequest struct {
	HEnum   mststs.ENUM_HANDLE
	State   int32
	BInvert ndr.BOOL
}

func (*rpcFilterByStateRequest) Opnum() uint16 { return TermSrvEnumeration.OpnumRpcFilterByState }

// rpcFilterByStateResponse carries the [out] parameters and return value of RpcFilterByState.
type rpcFilterByStateResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcFilterByState calls RpcFilterByState (opnum 2) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcFilterByState(rpc ndr.Invoker, hEnum mststs.ENUM_HANDLE, state int32, bInvert ndr.BOOL) (err error) {
	req := &rpcFilterByStateRequest{
		HEnum:   hEnum,
		State:   state,
		BInvert: bInvert,
	}
	var resp rpcFilterByStateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcFilterByState: %w", err)
		return
	}
	if uint32(resp.Status) != TermSrvEnumeration.StatusSuccess {
		err = fmt.Errorf("RpcFilterByState failed: %s", TermSrvEnumeration.StatusString(uint32(resp.Status)))
	}
	return
}
