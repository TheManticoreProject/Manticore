package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcAddPrinterRequest carries the [in] parameters of RpcAddPrinter.
type rpcAddPrinterRequest struct {
	PName              *ndr.WSTR `ndr:"unique"`
	PPrinterContainer  structures.PRINTER_CONTAINER
	PDevModeContainer  structures.DEVMODE_CONTAINER
	PSecurityContainer structures.SECURITY_CONTAINER
}

func (*rpcAddPrinterRequest) Opnum() uint16 { return winspool.OpnumRpcAddPrinter }

// rpcAddPrinterResponse carries the [out] parameters and return value of RpcAddPrinter.
type rpcAddPrinterResponse struct {
	PHandle structures.PRINTER_HANDLE
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcAddPrinter calls RpcAddPrinter (opnum 5) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcAddPrinter(rpc ndr.Invoker, pName *ndr.WSTR, pPrinterContainer structures.PRINTER_CONTAINER, pDevModeContainer structures.DEVMODE_CONTAINER, pSecurityContainer structures.SECURITY_CONTAINER) (PHandle structures.PRINTER_HANDLE, err error) {
	req := &rpcAddPrinterRequest{
		PName:              pName,
		PPrinterContainer:  pPrinterContainer,
		PDevModeContainer:  pDevModeContainer,
		PSecurityContainer: pSecurityContainer,
	}
	var resp rpcAddPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAddPrinter: %w", err)
		return
	}
	PHandle = resp.PHandle
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcAddPrinter failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
