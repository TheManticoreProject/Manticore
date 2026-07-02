package functions

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcSetPrinterDataExRequest carries the [in] parameters of RpcSetPrinterDataEx.
type rpcSetPrinterDataExRequest struct {
	HPrinter   structures.PRINTER_HANDLE
	PKeyName   ndr.WSTR
	PValueName ndr.WSTR
	Type       ndr.DWORD
	PData      []uint8 `ndr:"ref,size_is=CbData"`
	CbData     ndr.DWORD
}

func (*rpcSetPrinterDataExRequest) Opnum() uint16 { return winspool.OpnumRpcSetPrinterDataEx }

// rpcSetPrinterDataExResponse carries the [out] parameters and return value of RpcSetPrinterDataEx.
type rpcSetPrinterDataExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcSetPrinterDataEx calls RpcSetPrinterDataEx (opnum 77) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcSetPrinterDataEx(rpc ndr.Invoker, hPrinter structures.PRINTER_HANDLE, pKeyName ndr.WSTR, pValueName ndr.WSTR, type_ ndr.DWORD, pData []uint8, cbData ndr.DWORD) (err error) {
	req := &rpcSetPrinterDataExRequest{
		HPrinter:   hPrinter,
		PKeyName:   pKeyName,
		PValueName: pValueName,
		Type:       type_,
		PData:      pData,
		CbData:     cbData,
	}
	var resp rpcSetPrinterDataExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcSetPrinterDataEx: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcSetPrinterDataEx failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
