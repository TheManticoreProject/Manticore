package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcGetPrinterDataRequest carries the [in] parameters of RpcGetPrinterData.
type rpcGetPrinterDataRequest struct {
	HPrinter   structures.PRINTER_HANDLE
	PValueName ndr.WSTR
	NSize      ndr.DWORD
}

func (*rpcGetPrinterDataRequest) Opnum() uint16 { return winspool.OpnumRpcGetPrinterData }

// rpcGetPrinterDataResponse carries the [out] parameters and return value of RpcGetPrinterData.
type rpcGetPrinterDataResponse struct {
	PType     ndr.DWORD
	PData     []uint8 `ndr:"ref,size_is=NSize"`
	PcbNeeded ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcGetPrinterData calls RpcGetPrinterData (opnum 26) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcGetPrinterData(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, pValueName ndr.WSTR, nSize ndr.DWORD) (PType ndr.DWORD, PData []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcGetPrinterDataRequest{
		HPrinter:   hPrinter,
		PValueName: pValueName,
		NSize:      nSize,
	}
	var resp rpcGetPrinterDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetPrinterData: %w", err)
		return
	}
	PType = resp.PType
	PData = resp.PData
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcGetPrinterData failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
