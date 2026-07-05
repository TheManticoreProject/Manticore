package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationGetMachinePolicyRequest carries the [in] parameters of RpcWinStationGetMachinePolicy.
type rpcWinStationGetMachinePolicyRequest struct {
	HServer    mststs.SERVER_HANDLE
	PPolicy    []uint8 `ndr:"ref,size_is=BufferSize"`
	BufferSize ndr.DWORD
}

func (*rpcWinStationGetMachinePolicyRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationGetMachinePolicy
}

// rpcWinStationGetMachinePolicyResponse carries the [out] parameters and return value of RpcWinStationGetMachinePolicy.
type rpcWinStationGetMachinePolicyResponse struct {
	PPolicy []uint8   `ndr:"ref,size_is=BufferSize"`
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationGetMachinePolicy calls RpcWinStationGetMachinePolicy (opnum 62) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationGetMachinePolicy(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, pPolicy []uint8, bufferSize ndr.DWORD) (PPolicy []uint8, err error) {
	req := &rpcWinStationGetMachinePolicyRequest{
		HServer:    hServer,
		PPolicy:    pPolicy,
		BufferSize: bufferSize,
	}
	var resp rpcWinStationGetMachinePolicyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationGetMachinePolicy: %w", err)
		return
	}
	PPolicy = resp.PPolicy
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationGetMachinePolicy failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
