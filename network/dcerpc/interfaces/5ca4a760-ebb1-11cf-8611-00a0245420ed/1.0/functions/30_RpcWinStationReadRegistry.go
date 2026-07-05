package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationReadRegistryRequest carries the [in] parameters of RpcWinStationReadRegistry.
type rpcWinStationReadRegistryRequest struct {
	HServer mststs.SERVER_HANDLE
}

func (*rpcWinStationReadRegistryRequest) Opnum() uint16 { return IcaApi.OpnumRpcWinStationReadRegistry }

// rpcWinStationReadRegistryResponse carries the [out] parameters and return value of RpcWinStationReadRegistry.
type rpcWinStationReadRegistryResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationReadRegistry calls RpcWinStationReadRegistry (opnum 30) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationReadRegistry(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationReadRegistryRequest{
		HServer: hServer,
	}
	var resp rpcWinStationReadRegistryResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationReadRegistry: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationReadRegistry failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
