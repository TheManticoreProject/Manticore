package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationDisconnectRequest carries the [in] parameters of RpcWinStationDisconnect.
type rpcWinStationDisconnectRequest struct {
	HServer mststs.SERVER_HANDLE
	LogonId ndr.DWORD
	BWait   bool
}

func (*rpcWinStationDisconnectRequest) Opnum() uint16 { return IcaApi.OpnumRpcWinStationDisconnect }

// rpcWinStationDisconnectResponse carries the [out] parameters and return value of RpcWinStationDisconnect.
type rpcWinStationDisconnectResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationDisconnect calls RpcWinStationDisconnect (opnum 13) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationDisconnect(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, logonId ndr.DWORD, bWait bool) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationDisconnectRequest{
		HServer: hServer,
		LogonId: logonId,
		BWait:   bWait,
	}
	var resp rpcWinStationDisconnectResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationDisconnect: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationDisconnect failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
