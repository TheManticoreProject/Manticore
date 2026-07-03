package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcOpenPrinterExRequest carries the [in] parameters of RpcOpenPrinterEx.
type rpcOpenPrinterExRequest struct {
	PPrinterName      *ndr.WSTR `ndr:"unique"`
	PDatatype         *ndr.WSTR `ndr:"unique"`
	PDevModeContainer msrprn.DEVMODE_CONTAINER
	AccessRequired    ndr.DWORD
	PClientInfo       msrprn.SPLCLIENT_CONTAINER
}

func (*rpcOpenPrinterExRequest) Opnum() uint16 { return winspool.OpnumRpcOpenPrinterEx }

// rpcOpenPrinterExResponse carries the [out] parameters and return value of RpcOpenPrinterEx.
type rpcOpenPrinterExResponse struct {
	PHandle msrprn.PRINTER_HANDLE
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcOpenPrinterEx calls RpcOpenPrinterEx (opnum 69) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcOpenPrinterEx(rpc ndr.Invoker, pPrinterName *ndr.WSTR, pDatatype *ndr.WSTR, pDevModeContainer msrprn.DEVMODE_CONTAINER, accessRequired ndr.DWORD, pClientInfo msrprn.SPLCLIENT_CONTAINER) (PHandle msrprn.PRINTER_HANDLE, err error) {
	req := &rpcOpenPrinterExRequest{
		PPrinterName:      pPrinterName,
		PDatatype:         pDatatype,
		PDevModeContainer: pDevModeContainer,
		AccessRequired:    accessRequired,
		PClientInfo:       pClientInfo,
	}
	var resp rpcOpenPrinterExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcOpenPrinterEx: %w", err)
		return
	}
	PHandle = resp.PHandle
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcOpenPrinterEx failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
