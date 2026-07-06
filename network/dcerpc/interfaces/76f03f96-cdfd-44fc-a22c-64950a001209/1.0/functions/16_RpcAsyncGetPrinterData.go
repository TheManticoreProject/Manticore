package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncGetPrinterDataRequest carries the [in] parameters of RpcAsyncGetPrinterData.
type rpcAsyncGetPrinterDataRequest struct {
	HPrinter   mspar.PRINTER_HANDLE
	PValueName ndr.WSTR
	NSize      ndr.DWORD
}

func (*rpcAsyncGetPrinterDataRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncGetPrinterData
}

// rpcAsyncGetPrinterDataResponse carries the [out] parameters and return value of RpcAsyncGetPrinterData.
type rpcAsyncGetPrinterDataResponse struct {
	PType     ndr.DWORD
	PData     []uint8 `ndr:"ref,size_is=NSize"`
	PcbNeeded ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcAsyncGetPrinterData calls RpcAsyncGetPrinterData (opnum 16) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncGetPrinterData(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pValueName ndr.WSTR, nSize ndr.DWORD) (PType ndr.DWORD, PData []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcAsyncGetPrinterDataRequest{
		HPrinter:   hPrinter,
		PValueName: pValueName,
		NSize:      nSize,
	}
	var resp rpcAsyncGetPrinterDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncGetPrinterData: %w", err)
		return
	}
	PType = resp.PType
	PData = resp.PData
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncGetPrinterData failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
