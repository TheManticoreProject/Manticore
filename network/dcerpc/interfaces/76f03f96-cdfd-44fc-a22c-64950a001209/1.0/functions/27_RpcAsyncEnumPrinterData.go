package functions

import (
	"fmt"

	IRemoteWinspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/76f03f96-cdfd-44fc-a22c-64950a001209/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspar "github.com/TheManticoreProject/Manticore/windows/protocols/ms-par"
)

// rpcAsyncEnumPrinterDataRequest carries the [in] parameters of RpcAsyncEnumPrinterData.
type rpcAsyncEnumPrinterDataRequest struct {
	HPrinter    mspar.PRINTER_HANDLE
	DwIndex     ndr.DWORD
	CbValueName ndr.DWORD
	CbData      ndr.DWORD
}

func (*rpcAsyncEnumPrinterDataRequest) Opnum() uint16 {
	return IRemoteWinspool.OpnumRpcAsyncEnumPrinterData
}

// rpcAsyncEnumPrinterDataResponse carries the [out] parameters and return value of RpcAsyncEnumPrinterData.
type rpcAsyncEnumPrinterDataResponse struct {
	PValueName   []uint16 `ndr:"ref,size_is=CbValueName"`
	PcbValueName ndr.DWORD
	PType        ndr.DWORD
	PData        []uint8 `ndr:"ref,size_is=CbData"`
	PcbData      ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// RpcAsyncEnumPrinterData calls RpcAsyncEnumPrinterData (opnum 27) ([MS-PAR] — verify the parameter
// modeling and status handling).
func RpcAsyncEnumPrinterData(rpc ndr.Invoker, hPrinter mspar.PRINTER_HANDLE, dwIndex ndr.DWORD, cbValueName ndr.DWORD, cbData ndr.DWORD) (PValueName []uint16, PcbValueName ndr.DWORD, PType ndr.DWORD, PData []uint8, PcbData ndr.DWORD, err error) {
	req := &rpcAsyncEnumPrinterDataRequest{
		HPrinter:    hPrinter,
		DwIndex:     dwIndex,
		CbValueName: cbValueName,
		CbData:      cbData,
	}
	var resp rpcAsyncEnumPrinterDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcAsyncEnumPrinterData: %w", err)
		return
	}
	PValueName = resp.PValueName
	PcbValueName = resp.PcbValueName
	PType = resp.PType
	PData = resp.PData
	PcbData = resp.PcbData
	if uint32(resp.Status) != IRemoteWinspool.StatusSuccess {
		err = fmt.Errorf("RpcAsyncEnumPrinterData failed: %s", IRemoteWinspool.StatusString(uint32(resp.Status)))
	}
	return
}
