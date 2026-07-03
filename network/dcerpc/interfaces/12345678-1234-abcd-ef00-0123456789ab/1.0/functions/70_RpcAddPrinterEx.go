package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcAddPrinterExRequest carries the [in] parameters of RpcAddPrinterEx.
type rpcAddPrinterExRequest struct {
	PName              *ndr.WSTR `ndr:"unique"`
	PPrinterContainer  msrprn.PRINTER_CONTAINER
	PDevModeContainer  msrprn.DEVMODE_CONTAINER
	PSecurityContainer msrprn.SECURITY_CONTAINER
	PClientInfo        msrprn.SPLCLIENT_CONTAINER
}

func (*rpcAddPrinterExRequest) Opnum() uint16 { return winspool.OpnumRpcAddPrinterEx }

// rpcAddPrinterExResponse carries the [out] parameters and return value of RpcAddPrinterEx.
type rpcAddPrinterExResponse struct {
	PHandle msrprn.PRINTER_HANDLE
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcAddPrinterEx calls RpcAddPrinterEx (opnum 70) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcAddPrinterEx(rpc ndr.Invoker, pName *ndr.WSTR, pPrinterContainer msrprn.PRINTER_CONTAINER, pDevModeContainer msrprn.DEVMODE_CONTAINER, pSecurityContainer msrprn.SECURITY_CONTAINER, pClientInfo msrprn.SPLCLIENT_CONTAINER) (PHandle msrprn.PRINTER_HANDLE, err error) {
	req := &rpcAddPrinterExRequest{
		PName:              pName,
		PPrinterContainer:  pPrinterContainer,
		PDevModeContainer:  pDevModeContainer,
		PSecurityContainer: pSecurityContainer,
		PClientInfo:        pClientInfo,
	}
	var resp rpcAddPrinterExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAddPrinterEx: %w", err)
		return
	}
	PHandle = resp.PHandle
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcAddPrinterEx failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
