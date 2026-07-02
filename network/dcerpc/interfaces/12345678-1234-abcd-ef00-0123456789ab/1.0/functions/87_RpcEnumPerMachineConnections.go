package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcEnumPerMachineConnectionsRequest carries the [in] parameters of RpcEnumPerMachineConnections.
type rpcEnumPerMachineConnectionsRequest struct {
	PServer      *ndr.WSTR `ndr:"unique"`
	PPrinterEnum []uint8   `ndr:"unique,size_is=CbBuf"`
	CbBuf        ndr.DWORD
}

func (*rpcEnumPerMachineConnectionsRequest) Opnum() uint16 {
	return winspool.OpnumRpcEnumPerMachineConnections
}

// rpcEnumPerMachineConnectionsResponse carries the [out] parameters and return value of RpcEnumPerMachineConnections.
type rpcEnumPerMachineConnectionsResponse struct {
	PPrinterEnum []uint8 `ndr:"unique,size_is=CbBuf"`
	PcbNeeded    ndr.DWORD
	PcReturned   ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// RpcEnumPerMachineConnections calls RpcEnumPerMachineConnections (opnum 87) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcEnumPerMachineConnections(rpc ndr.Invoker, pServer *ndr.WSTR, pPrinterEnum []uint8, cbBuf ndr.DWORD) (PPrinterEnum []uint8, PcbNeeded ndr.DWORD, PcReturned ndr.DWORD, err error) {
	req := &rpcEnumPerMachineConnectionsRequest{
		PServer:      pServer,
		PPrinterEnum: pPrinterEnum,
		CbBuf:        cbBuf,
	}
	var resp rpcEnumPerMachineConnectionsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcEnumPerMachineConnections: %w", err)
		return
	}
	PPrinterEnum = resp.PPrinterEnum
	PcbNeeded = resp.PcbNeeded
	PcReturned = resp.PcReturned
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcEnumPerMachineConnections failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
