package functions

import (
	"fmt"

	TermSrvEnumeration "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/88143fd0-c28d-4b2b-8fef-8d882f6a9390/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcFilterBySessionTypeRequest carries the [in] parameters of RpcFilterBySessionType.
type rpcFilterBySessionTypeRequest struct {
	HEnum        mststs.ENUM_HANDLE
	PSessionType guid.GUID
}

func (*rpcFilterBySessionTypeRequest) Opnum() uint16 {
	return TermSrvEnumeration.OpnumRpcFilterBySessionType
}

// rpcFilterBySessionTypeResponse carries the [out] parameters and return value of RpcFilterBySessionType.
type rpcFilterBySessionTypeResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcFilterBySessionType calls RpcFilterBySessionType (opnum 6) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcFilterBySessionType(rpc ndr.Invoker, hEnum mststs.ENUM_HANDLE, pSessionType guid.GUID) (err error) {
	req := &rpcFilterBySessionTypeRequest{
		HEnum:        hEnum,
		PSessionType: pSessionType,
	}
	var resp rpcFilterBySessionTypeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcFilterBySessionType: %w", err)
		return
	}
	if uint32(resp.Status) != TermSrvEnumeration.StatusSuccess {
		err = fmt.Errorf("RpcFilterBySessionType failed: %s", TermSrvEnumeration.StatusString(uint32(resp.Status)))
	}
	return
}
