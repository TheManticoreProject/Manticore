package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationCloseServerRequest carries the [in] parameters of RpcWinStationCloseServer.
type rpcWinStationCloseServerRequest struct {
	HServer mststs.SERVER_HANDLE
}

func (*rpcWinStationCloseServerRequest) Opnum() uint16 { return IcaApi.OpnumRpcWinStationCloseServer }

// rpcWinStationCloseServerResponse carries the [out] parameters and return value of RpcWinStationCloseServer.
type rpcWinStationCloseServerResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationCloseServer calls RpcWinStationCloseServer (opnum 1) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationCloseServer(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationCloseServerRequest{
		HServer: hServer,
	}
	var resp rpcWinStationCloseServerResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationCloseServer: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationCloseServer failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
