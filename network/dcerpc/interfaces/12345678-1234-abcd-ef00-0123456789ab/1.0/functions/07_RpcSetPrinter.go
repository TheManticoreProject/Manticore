package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcSetPrinterRequest carries the [in] parameters of RpcSetPrinter.
type rpcSetPrinterRequest struct {
	HPrinter           structures.PRINTER_HANDLE
	PPrinterContainer  structures.PRINTER_CONTAINER
	PDevModeContainer  structures.DEVMODE_CONTAINER
	PSecurityContainer structures.SECURITY_CONTAINER
	Command            ndr.DWORD
}

func (*rpcSetPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcSetPrinter }

// rpcSetPrinterResponse carries the [out] parameters and return value of RpcSetPrinter.
type rpcSetPrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcSetPrinter calls RpcSetPrinter (opnum 7) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcSetPrinter(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, pPrinterContainer structures.PRINTER_CONTAINER, pDevModeContainer structures.DEVMODE_CONTAINER, pSecurityContainer structures.SECURITY_CONTAINER, command ndr.DWORD) (err error) {
	req := &rpcSetPrinterRequest{
		HPrinter:           hPrinter,
		PPrinterContainer:  pPrinterContainer,
		PDevModeContainer:  pDevModeContainer,
		PSecurityContainer: pSecurityContainer,
		Command:            command,
	}
	var resp rpcSetPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcSetPrinter: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcSetPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
