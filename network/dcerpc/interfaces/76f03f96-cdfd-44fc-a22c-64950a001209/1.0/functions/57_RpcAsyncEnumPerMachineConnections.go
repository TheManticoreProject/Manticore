package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncEnumPerMachineConnectionsRequest carries the [in] parameters of RpcAsyncEnumPerMachineConnections.
type rpcAsyncEnumPerMachineConnectionsRequest struct {
	PServer      *ndr.WSTR `ndr:"unique"`
	PPrinterEnum []uint8   `ndr:"ref,size_is=CbBuf"`
	CbBuf        ndr.DWORD
}

func (*rpcAsyncEnumPerMachineConnectionsRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncEnumPerMachineConnections
}

// rpcAsyncEnumPerMachineConnectionsResponse carries the [out] parameters and return value of RpcAsyncEnumPerMachineConnections.
type rpcAsyncEnumPerMachineConnectionsResponse struct {
	PPrinterEnum []uint8 `ndr:"ref,size_is=CbBuf"`
	PcbNeeded    ndr.DWORD
	PcReturned   ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// RpcAsyncEnumPerMachineConnections calls RpcAsyncEnumPerMachineConnections (opnum 57) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEnumPerMachineConnections(rpc ndr.Invoker, pServer *ndr.WSTR, pPrinterEnum []uint8, cbBuf ndr.DWORD) (PPrinterEnum []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcAsyncEnumPerMachineConnectionsRequest{
		PServer:      pServer,
		PPrinterEnum: pPrinterEnum,
		CbBuf:        cbBuf,
	}
	var resp rpcAsyncEnumPerMachineConnectionsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEnumPerMachineConnections: %w", err)
		return
	}
	PPrinterEnum = resp.PPrinterEnum
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEnumPerMachineConnections failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
