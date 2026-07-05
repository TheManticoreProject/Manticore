package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationIsHelpAssistantSessionRequest carries the [in] parameters of RpcWinStationIsHelpAssistantSession.
type rpcWinStationIsHelpAssistantSessionRequest struct {
	HServer   mststs.SERVER_HANDLE
	SessionId ndr.DWORD
}

func (*rpcWinStationIsHelpAssistantSessionRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationIsHelpAssistantSession
}

// rpcWinStationIsHelpAssistantSessionResponse carries the [out] parameters and return value of RpcWinStationIsHelpAssistantSession.
type rpcWinStationIsHelpAssistantSessionResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationIsHelpAssistantSession calls RpcWinStationIsHelpAssistantSession (opnum 61) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationIsHelpAssistantSession(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, sessionId ndr.DWORD) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationIsHelpAssistantSessionRequest{
		HServer:   hServer,
		SessionId: sessionId,
	}
	var resp rpcWinStationIsHelpAssistantSessionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationIsHelpAssistantSession: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationIsHelpAssistantSession failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
