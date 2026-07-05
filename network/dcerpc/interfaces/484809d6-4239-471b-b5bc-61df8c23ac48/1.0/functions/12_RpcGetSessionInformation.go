package functions

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetSessionInformationRequest carries the [in] parameters of RpcGetSessionInformation.
type rpcGetSessionInformationRequest struct {
	SessionId int32
}

func (*rpcGetSessionInformationRequest) Opnum() uint16 {
	return TermSrvSession.OpnumRpcGetSessionInformation
}

// rpcGetSessionInformationResponse carries the [out] parameters and return value of RpcGetSessionInformation.
type rpcGetSessionInformationResponse struct {
	PSessionInfo mststs.LSMSESSIONINFORMATION
	Status       ndr.DWORD `ndr:"retval"`
}

// RpcGetSessionInformation calls RpcGetSessionInformation (opnum 12) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetSessionInformation(rpc ndr.Invoker, sessionId int32) (PSessionInfo mststs.LSMSESSIONINFORMATION, err error) {
	req := &rpcGetSessionInformationRequest{
		SessionId: sessionId,
	}
	var resp rpcGetSessionInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetSessionInformation: %w", err)
		return
	}
	PSessionInfo = resp.PSessionInfo
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcGetSessionInformation failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
