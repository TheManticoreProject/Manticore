package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcOpenPrinterRequest carries the [in] parameters of RpcOpenPrinter.
type rpcOpenPrinterRequest struct {
	PPrinterName      *ndr.WSTR `ndr:"unique"`
	PDatatype         *ndr.WSTR `ndr:"unique"`
	PDevModeContainer structures.DEVMODE_CONTAINER
	AccessRequired    ndr.DWORD
}

func (*rpcOpenPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcOpenPrinter }

// rpcOpenPrinterResponse carries the [out] parameters and return value of RpcOpenPrinter.
type rpcOpenPrinterResponse struct {
	PHandle structures.PRINTER_HANDLE
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcOpenPrinter calls RpcOpenPrinter (opnum 1) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcOpenPrinter(rpc ndr.Invoker, pPrinterName *ndr.WSTR, pDatatype *ndr.WSTR, pDevModeContainer structures.DEVMODE_CONTAINER, accessRequired ndr.DWORD) (PHandle structures.PRINTER_HANDLE, err error) {
	req := &rpcOpenPrinterRequest{
		PPrinterName:      pPrinterName,
		PDatatype:         pDatatype,
		PDevModeContainer: pDevModeContainer,
		AccessRequired:    accessRequired,
	}
	var resp rpcOpenPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcOpenPrinter: %w", err)
		return
	}
	PHandle = resp.PHandle
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcOpenPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
