package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAddPerMachineConnectionRequest carries the [in] parameters of RpcAddPerMachineConnection.
type rpcAddPerMachineConnectionRequest struct {
	PServer      *ndr.WSTR `ndr:"unique"`
	PPrinterName ndr.WSTR
	PPrintServer ndr.WSTR
	PProvider    ndr.WSTR
}

func (*rpcAddPerMachineConnectionRequest) Opnum() uint16 {
	return winspool.OpnumRpcAddPerMachineConnection
}

// rpcAddPerMachineConnectionResponse carries the [out] parameters and return value of RpcAddPerMachineConnection.
type rpcAddPerMachineConnectionResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAddPerMachineConnection calls RpcAddPerMachineConnection (opnum 85) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcAddPerMachineConnection(rpc ndr.Invoker, pServer *ndr.WSTR, pPrinterName ndr.WSTR, pPrintServer ndr.WSTR, pProvider ndr.WSTR) (err error) {
	req := &rpcAddPerMachineConnectionRequest{
		PServer:      pServer,
		PPrinterName: pPrinterName,
		PPrintServer: pPrintServer,
		PProvider:    pProvider,
	}
	var resp rpcAddPerMachineConnectionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAddPerMachineConnection: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcAddPerMachineConnection failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
