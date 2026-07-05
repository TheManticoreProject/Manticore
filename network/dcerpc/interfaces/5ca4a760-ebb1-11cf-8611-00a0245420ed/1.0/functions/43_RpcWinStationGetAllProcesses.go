package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationGetAllProcessesRequest carries the [in] parameters of RpcWinStationGetAllProcesses.
type rpcWinStationGetAllProcessesRequest struct {
	HServer            mststs.SERVER_HANDLE
	Level              ndr.DWORD
	PNumberOfProcesses ndr.DWORD
}

func (*rpcWinStationGetAllProcessesRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationGetAllProcesses
}

// rpcWinStationGetAllProcessesResponse carries the [out] parameters and return value of RpcWinStationGetAllProcesses.
type rpcWinStationGetAllProcessesResponse struct {
	PResult              ndr.DWORD
	PNumberOfProcesses   ndr.DWORD
	PpTsAllProcessesInfo []mststs.TS_ALL_PROCESSES_INFO `ndr:"unique,conformant"`
	Status               ndr.DWORD                      `ndr:"retval"`
}

// RpcWinStationGetAllProcesses calls RpcWinStationGetAllProcesses (opnum 43) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationGetAllProcesses(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, level ndr.DWORD, pNumberOfProcesses ndr.DWORD) (PResult ndr.DWORD, PNumberOfProcesses ndr.DWORD, PpTsAllProcessesInfo []mststs.TS_ALL_PROCESSES_INFO, err error) {
	req := &rpcWinStationGetAllProcessesRequest{
		HServer:            hServer,
		Level:              level,
		PNumberOfProcesses: pNumberOfProcesses,
	}
	var resp rpcWinStationGetAllProcessesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationGetAllProcesses: %w", err)
		return
	}
	PResult = resp.PResult
	PNumberOfProcesses = resp.PNumberOfProcesses
	PpTsAllProcessesInfo = resp.PpTsAllProcessesInfo
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationGetAllProcesses failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
