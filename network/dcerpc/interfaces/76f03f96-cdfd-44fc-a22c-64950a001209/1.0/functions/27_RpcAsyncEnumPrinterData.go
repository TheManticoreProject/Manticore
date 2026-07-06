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
