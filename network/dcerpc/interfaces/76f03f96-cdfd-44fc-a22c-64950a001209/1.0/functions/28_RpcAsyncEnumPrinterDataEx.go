package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncEnumPrinterDataExRequest carries the [in] parameters of RpcAsyncEnumPrinterDataEx.
type rpcAsyncEnumPrinterDataExRequest struct {
	HPrinter     mspar.PRINTER_HANDLE
	PKeyName     ndr.WSTR
	CbEnumValues ndr.DWORD
}

func (*rpcAsyncEnumPrinterDataExRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncEnumPrinterDataEx
}

// rpcAsyncEnumPrinterDataExResponse carries the [out] parameters and return value of RpcAsyncEnumPrinterDataEx.
type rpcAsyncEnumPrinterDataExResponse struct {
	PEnumValues   []uint8 `ndr:"ref,size_is=CbEnumValues"`
	PcbEnumValues ndr.DWORD
	PnEnumValues  ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// RpcAsyncEnumPrinterDataEx calls RpcAsyncEnumPrinterDataEx (opnum 28) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEnumPrinterDataEx(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pKeyName ndr.WSTR, cbEnumValues ndr.DWORD) (PEnumValues []uint8, PcbEnumValues ndr.DWORD, PnEnumValues ndr.DWORD, err error) {
	req := &rpcAsyncEnumPrinterDataExRequest{
		HPrinter:     hPrinter,
		PKeyName:     pKeyName,
		CbEnumValues: cbEnumValues,
	}
	var resp rpcAsyncEnumPrinterDataExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEnumPrinterDataEx: %w", err)
		return
	}
	PEnumValues = resp.PEnumValues
	PcbEnumValues = resp.PcbEnumValues
	PnEnumValues = resp.PnEnumValues
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEnumPrinterDataEx failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
