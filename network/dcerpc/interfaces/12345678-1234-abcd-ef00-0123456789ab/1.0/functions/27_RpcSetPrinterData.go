package functions

// IDL source: [MS-RPRN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rprn/e8f9dad8-d114-41cc-9a52-fc927e908cf4
// A fetched copy is kept at ms-rprn.idl in the interface directory.

import (
	"fmt"

	winspool "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-0123456789ab/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrprn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rprn"
)

// rpcSetPrinterDataRequest carries the [in] parameters of RpcSetPrinterData.
type rpcSetPrinterDataRequest struct {
	HPrinter   msrprn.PRINTER_HANDLE
	PValueName ndr.WSTR
	Type       ndr.DWORD
	PData      []uint8 `ndr:"ref,size_is=CbData"`
	CbData     ndr.DWORD
}

func (*rpcSetPrinterDataRequest) Opnum() uint16 { return winspool.OpnumRpcSetPrinterData }

// rpcSetPrinterDataResponse carries the [out] parameters and return value of RpcSetPrinterData.
type rpcSetPrinterDataResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcSetPrinterData calls RpcSetPrinterData (opnum 27) ([MS-RPRN] — verify the parameter
// modeling and status handling).
func RpcSetPrinterData(rpc ndr.Invoker, hPrinter msrprn.PRINTER_HANDLE, pValueName ndr.WSTR, type_ ndr.DWORD, pData []uint8, cbData ndr.DWORD) (err error) {
	req := &rpcSetPrinterDataRequest{
		HPrinter:   hPrinter,
		PValueName: pValueName,
		Type:       type_,
		PData:      pData,
		CbData:     cbData,
	}
	var resp rpcSetPrinterDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcSetPrinterData: %w", err)
		return
	}
	if uint32(resp.Status) != winspool.StatusSuccess {
		err = fmt.Errorf("RpcSetPrinterData failed: %s", winspool.StatusString(uint32(resp.Status)))
	}
	return
}
