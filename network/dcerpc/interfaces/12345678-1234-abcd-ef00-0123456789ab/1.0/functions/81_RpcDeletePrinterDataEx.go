package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcDeletePrinterDataExRequest carries the [in] parameters of RpcDeletePrinterDataEx.
type rpcDeletePrinterDataExRequest struct {
	HPrinter   msrprn.PRINTER_HANDLE
	PKeyName   ndr.WSTR
	PValueName ndr.WSTR
}

func (*rpcDeletePrinterDataExRequest) Opnum() uint16 { return winspool.OpnumRpcDeletePrinterDataEx }

// rpcDeletePrinterDataExResponse carries the [out] parameters and return value of RpcDeletePrinterDataEx.
type rpcDeletePrinterDataExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcDeletePrinterDataEx calls RpcDeletePrinterDataEx (opnum 81) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcDeletePrinterDataEx(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, pKeyName ndr.WSTR, pValueName ndr.WSTR) (err error) {
	req := &rpcDeletePrinterDataExRequest{
		HPrinter:   hPrinter,
		PKeyName:   pKeyName,
		PValueName: pValueName,
	}
	var resp rpcDeletePrinterDataExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcDeletePrinterDataEx: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcDeletePrinterDataEx failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
