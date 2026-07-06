package functions

// IDL source: [MS-PAR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-par/d81865df-838d-4c13-a705-d41ee24890de
// A fetched copy is kept at ms-par.idl in the interface directory.

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncOpenPrinterRequest carries the [in] parameters of RpcAsyncOpenPrinter.
type rpcAsyncOpenPrinterRequest struct {
	PPrinterName      *ndr.WSTR `ndr:"unique"`
	PDatatype         *ndr.WSTR `ndr:"unique"`
	PDevModeContainer mspar.DEVMODE_CONTAINER
	AccessRequired    ndr.DWORD
	PClientInfo       mspar.SPLCLIENT_CONTAINER
}

func (*rpcAsyncOpenPrinterRequest) Opnum() uint16 { return IRemoteWinspool.OpnumRpcAsyncOpenPrinter }

// rpcAsyncOpenPrinterResponse carries the [out] parameters and return value of RpcAsyncOpenPrinter.
type rpcAsyncOpenPrinterResponse struct {
	PHandle mspar.PRINTER_HANDLE
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcAsyncOpenPrinter calls RpcAsyncOpenPrinter (opnum 0) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncOpenPrinter(rpc ndr.Invoker, pPrinterName *ndr.WSTR, pDatatype *ndr.WSTR, pDevModeContainer mspar.DEVMODE_CONTAINER, accessRequired ndr.DWORD, pClientInfo mspar.SPLCLIENT_CONTAINER) (PHandle mspar.PRINTER_HANDLE, err error) {
	req := &rpcAsyncOpenPrinterRequest{
		PPrinterName:      pPrinterName,
		PDatatype:         pDatatype,
		PDevModeContainer: pDevModeContainer,
		AccessRequired:    accessRequired,
		PClientInfo:       pClientInfo,
	}
	var resp rpcAsyncOpenPrinterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncOpenPrinter: %w", err)
		return
	}
	PHandle = resp.PHandle
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncOpenPrinter failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
