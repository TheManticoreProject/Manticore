package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationBreakPointRequest carries the [in] parameters of RpcWinStationBreakPoint.
type rpcWinStationBreakPointRequest struct {
	HServer    mststs.SERVER_HANDLE
	LogonId    ndr.DWORD
	KernelFlag bool
}

func (*rpcWinStationBreakPointRequest) Opnum() uint16 { return IcaApi.OpnumRpcWinStationBreakPoint }

// rpcWinStationBreakPointResponse carries the [out] parameters and return value of RpcWinStationBreakPoint.
type rpcWinStationBreakPointResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationBreakPoint calls RpcWinStationBreakPoint (opnum 29) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationBreakPoint(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, logonId ndr.DWORD, kernelFlag bool) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationBreakPointRequest{
		HServer:    hServer,
		LogonId:    logonId,
		KernelFlag: kernelFlag,
	}
	var resp rpcWinStationBreakPointResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationBreakPoint: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationBreakPoint failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
