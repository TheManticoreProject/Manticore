package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAsyncAddPerMachineConnectionRequest carries the [in] parameters of RpcAsyncAddPerMachineConnection.
type rpcAsyncAddPerMachineConnectionRequest struct {
	PServer      *ndr.WSTR `ndr:"unique"`
	PPrinterName ndr.WSTR
	PPrintServer ndr.WSTR
	PProvider    ndr.WSTR
}

func (*rpcAsyncAddPerMachineConnectionRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncAddPerMachineConnection
}

// rpcAsyncAddPerMachineConnectionResponse carries the [out] parameters and return value of RpcAsyncAddPerMachineConnection.
type rpcAsyncAddPerMachineConnectionResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncAddPerMachineConnection calls RpcAsyncAddPerMachineConnection (opnum 55) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncAddPerMachineConnection(rpc ndr.Invoker, pServer *ndr.WSTR, pPrinterName ndr.WSTR, pPrintServer ndr.WSTR, pProvider ndr.WSTR) (err error) {
	req := &rpcAsyncAddPerMachineConnectionRequest{
		PServer:      pServer,
		PPrinterName: pPrinterName,
		PPrintServer: pPrintServer,
		PProvider:    pProvider,
	}
	var resp rpcAsyncAddPerMachineConnectionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncAddPerMachineConnection: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncAddPerMachineConnection failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
