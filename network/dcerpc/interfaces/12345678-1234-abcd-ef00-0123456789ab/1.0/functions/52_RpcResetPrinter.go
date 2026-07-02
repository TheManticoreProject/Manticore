package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcResetPrinterRequest carries the [in] parameters of RpcResetPrinter.
type rpcResetPrinterRequest struct {
	HPrinter          structures.PRINTER_HANDLE
	PDatatype         *ndr.WSTR `ndr:"unique"`
	PDevModeContainer structures.DEVMODE_CONTAINER
}

func (*rpcResetPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcResetPrinter }

// rpcResetPrinterResponse carries the [out] parameters and return value of RpcResetPrinter.
type rpcResetPrinterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcResetPrinter calls RpcResetPrinter (opnum 52) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcResetPrinter(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, pDatatype *ndr.WSTR, pDevModeContainer structures.DEVMODE_CONTAINER) (err error) {
	req := &rpcResetPrinterRequest{
		HPrinter:          hPrinter,
		PDatatype:         pDatatype,
		PDevModeContainer: pDevModeContainer,
	}
	var resp rpcResetPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcResetPrinter: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcResetPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
