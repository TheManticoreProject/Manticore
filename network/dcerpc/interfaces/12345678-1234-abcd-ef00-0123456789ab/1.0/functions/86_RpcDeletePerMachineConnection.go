package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcDeletePerMachineConnectionRequest carries the [in] parameters of RpcDeletePerMachineConnection.
type rpcDeletePerMachineConnectionRequest struct {
	PServer      *ndr.WSTR `ndr:"unique"`
	PPrinterName ndr.WSTR
}

func (*rpcDeletePerMachineConnectionRequest) Opnum() uint16 {
	return winspool.OpnumRpcDeletePerMachineConnection
}

// rpcDeletePerMachineConnectionResponse carries the [out] parameters and return value of RpcDeletePerMachineConnection.
type rpcDeletePerMachineConnectionResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDeletePerMachineConnection calls RpcDeletePerMachineConnection (opnum 86) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeletePerMachineConnection(rpc ndr.Invoker, pServer *ndr.WSTR, pPrinterName ndr.WSTR) (err error) {
	req := &rpcDeletePerMachineConnectionRequest{
		PServer:      pServer,
		PPrinterName: pPrinterName,
	}
	var resp rpcDeletePerMachineConnectionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDeletePerMachineConnection: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcDeletePerMachineConnection failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
