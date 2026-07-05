package functions

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcGetSessionTypeRequest carries the [in] parameters of RpcGetSessionType.
type rpcGetSessionTypeRequest struct {
	SessionId int32
}

func (*rpcGetSessionTypeRequest) Opnum() uint16 { return TermSrvSession.OpnumRpcGetSessionType }

// rpcGetSessionTypeResponse carries the [out] parameters and return value of RpcGetSessionType.
type rpcGetSessionTypeResponse struct {
	PSessionType ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// RpcGetSessionType calls RpcGetSessionType (opnum 16) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetSessionType(rpc ndr.Invoker, sessionId int32) (PSessionType ndr.DWORD, err error) {
	req := &rpcGetSessionTypeRequest{
		SessionId: sessionId,
	}
	var resp rpcGetSessionTypeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetSessionType: %w", err)
		return
	}
	PSessionType = resp.PSessionType
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcGetSessionType failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
