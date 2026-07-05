package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationShadowStopRequest carries the [in] parameters of RpcWinStationShadowStop.
type rpcWinStationShadowStopRequest struct {
	HServer mststs.SERVER_HANDLE
	LogonId ndr.DWORD
	BWait   bool
}

func (*rpcWinStationShadowStopRequest) Opnum() uint16 { return IcaApi.OpnumRpcWinStationShadowStop }

// rpcWinStationShadowStopResponse carries the [out] parameters and return value of RpcWinStationShadowStop.
type rpcWinStationShadowStopResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationShadowStop calls RpcWinStationShadowStop (opnum 59) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationShadowStop(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, logonId ndr.DWORD, bWait bool) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationShadowStopRequest{
		HServer: hServer,
		LogonId: logonId,
		BWait:   bWait,
	}
	var resp rpcWinStationShadowStopResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationShadowStop: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationShadowStop failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
