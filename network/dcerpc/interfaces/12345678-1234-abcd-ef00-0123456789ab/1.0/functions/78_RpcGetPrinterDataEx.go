package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcGetPrinterDataExRequest carries the [in] parameters of RpcGetPrinterDataEx.
type rpcGetPrinterDataExRequest struct {
	HPrinter   msrprn.PRINTER_HANDLE
	PKeyName   ndr.WSTR
	PValueName ndr.WSTR
	NSize      ndr.DWORD
}

func (*rpcGetPrinterDataExRequest) Opnum() uint16 { return winspool.OpnumRpcGetPrinterDataEx }

// rpcGetPrinterDataExResponse carries the [out] parameters and return value of RpcGetPrinterDataEx.
type rpcGetPrinterDataExResponse struct {
	PType     ndr.DWORD
	PData     []uint8 `ndr:"ref,size_is=NSize"`
	PcbNeeded ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcGetPrinterDataEx calls RpcGetPrinterDataEx (opnum 78) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcGetPrinterDataEx(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, pKeyName ndr.WSTR, pValueName ndr.WSTR, nSize ndr.DWORD) (PType ndr.DWORD, PData []uint8, PcbNeeded ndr.DWORD, err error) {
	req := &rpcGetPrinterDataExRequest{
		HPrinter:   hPrinter,
		PKeyName:   pKeyName,
		PValueName: pValueName,
		NSize:      nSize,
	}
	var resp rpcGetPrinterDataExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetPrinterDataEx: %w", err)
		return
	}
	PType = resp.PType
	PData = resp.PData
	PcbNeeded = resp.PcbNeeded
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcGetPrinterDataEx failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
