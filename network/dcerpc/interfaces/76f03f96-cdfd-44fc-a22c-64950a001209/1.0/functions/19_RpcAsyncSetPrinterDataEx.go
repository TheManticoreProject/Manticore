package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncSetPrinterDataExRequest carries the [in] parameters of RpcAsyncSetPrinterDataEx.
type rpcAsyncSetPrinterDataExRequest struct {
	HPrinter   mspar.PRINTER_HANDLE
	PKeyName   ndr.WSTR
	PValueName ndr.WSTR
	Type       ndr.DWORD
	PData      []uint8 `ndr:"ref,size_is=CbData"`
	CbData     ndr.DWORD
}

func (*rpcAsyncSetPrinterDataExRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncSetPrinterDataEx
}

// rpcAsyncSetPrinterDataExResponse carries the [out] parameters and return value of RpcAsyncSetPrinterDataEx.
type rpcAsyncSetPrinterDataExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncSetPrinterDataEx calls RpcAsyncSetPrinterDataEx (opnum 19) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncSetPrinterDataEx(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pKeyName ndr.WSTR, pValueName ndr.WSTR, type_ ndr.DWORD, pData []uint8, cbData ndr.DWORD) (err error) {
	req := &rpcAsyncSetPrinterDataExRequest{
		HPrinter:   hPrinter,
		PKeyName:   pKeyName,
		PValueName: pValueName,
		Type:       type_,
		PData:      pData,
		CbData:     cbData,
	}
	var resp rpcAsyncSetPrinterDataExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncSetPrinterDataEx: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncSetPrinterDataEx failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
