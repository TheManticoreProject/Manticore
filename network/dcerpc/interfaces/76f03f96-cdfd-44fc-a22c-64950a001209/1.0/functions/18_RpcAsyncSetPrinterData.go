package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncSetPrinterDataRequest carries the [in] parameters of RpcAsyncSetPrinterData.
type rpcAsyncSetPrinterDataRequest struct {
	HPrinter   mspar.PRINTER_HANDLE
	PValueName ndr.WSTR
	Type       ndr.DWORD
	PData      []uint8 `ndr:"ref,size_is=CbData"`
	CbData     ndr.DWORD
}

func (*rpcAsyncSetPrinterDataRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncSetPrinterData
}

// rpcAsyncSetPrinterDataResponse carries the [out] parameters and return value of RpcAsyncSetPrinterData.
type rpcAsyncSetPrinterDataResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcAsyncSetPrinterData calls RpcAsyncSetPrinterData (opnum 18) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncSetPrinterData(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, pValueName ndr.WSTR, type_ ndr.DWORD, pData []uint8, cbData ndr.DWORD) (err error) {
	req := &rpcAsyncSetPrinterDataRequest{
		HPrinter:   hPrinter,
		PValueName: pValueName,
		Type:       type_,
		PData:      pData,
		CbData:     cbData,
	}
	var resp rpcAsyncSetPrinterDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncSetPrinterData: %w", err)
		return
	}
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncSetPrinterData failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
