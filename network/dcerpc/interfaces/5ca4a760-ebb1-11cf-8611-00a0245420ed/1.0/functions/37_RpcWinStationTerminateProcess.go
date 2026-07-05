package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationTerminateProcessRequest carries the [in] parameters of RpcWinStationTerminateProcess.
type rpcWinStationTerminateProcessRequest struct {
	HServer   mststs.SERVER_HANDLE
	ProcessId ndr.DWORD
	ExitCode  ndr.DWORD
}

func (*rpcWinStationTerminateProcessRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationTerminateProcess
}

// rpcWinStationTerminateProcessResponse carries the [out] parameters and return value of RpcWinStationTerminateProcess.
type rpcWinStationTerminateProcessResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationTerminateProcess calls RpcWinStationTerminateProcess (opnum 37) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationTerminateProcess(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, processId ndr.DWORD, exitCode ndr.DWORD) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationTerminateProcessRequest{
		HServer:   hServer,
		ProcessId: processId,
		ExitCode:  exitCode,
	}
	var resp rpcWinStationTerminateProcessResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationTerminateProcess: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationTerminateProcess failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
