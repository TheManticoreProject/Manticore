package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncGetPrinterDataExRequest carries the [in] parameters of RpcAsyncGetPrinterDataEx.
type rpcAsyncGetPrinterDataExRequest struct {
	HPrinter   mspar.PRINTER_HANDLE
	PKeyName   ndr.WSTR
	PValueName ndr.WSTR
	NSize      ndr.DWORD
}

func (*rpcAsyncGetPrinterDataExRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncGetPrinterDataEx
}

// rpcAsyncGetPrinterDataExResponse carries the [out] parameters and return value of RpcAsyncGetPrinterDataEx.
type rpcAsyncGetPrinterDataExResponse struct {
	PType     ndr.DWORD
	PData     []uint8 `ndr:"ref,size_is=NSize"`
	PcbNeeded ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcAsyncGetPrinterDataEx calls RpcAsyncGetPrinterDataEx (opnum 17) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncGetPrinterDataEx(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pKeyName ndr.WSTR, pValueName ndr.WSTR, nSize ndr.DWORD) (PType ndr.DWORD, PData []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcAsyncGetPrinterDataExRequest{
		HPrinter:   hPrinter,
		PKeyName:   pKeyName,
		PValueName: pValueName,
		NSize:      nSize,
	}
	var resp rpcAsyncGetPrinterDataExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncGetPrinterDataEx: %w", err)
		return
	}
	PType = resp.PType
	PData = resp.PData
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncGetPrinterDataEx failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
